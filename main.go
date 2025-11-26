package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	rtctokenbuilder2 "github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	rtmtokenbuilder2 "github.com/AgoraIO-Community/go-tokenbuilder/rtmtokenbuilder"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var appID, appCertificate string

// In-memory huddle storage
var huddleStore *HuddleStore

// Huddle represents a huddle room
type Huddle struct {
	ID           string    `json:"huddle_id"`
	ChannelName  string    `json:"channel_name"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	Participants []string  `json:"participants"`
}

// HuddleStore manages huddles in memory with thread-safe operations
type HuddleStore struct {
	mu      sync.RWMutex
	huddles map[string]*Huddle
}

// NewHuddleStore creates a new in-memory huddle store
func NewHuddleStore() *HuddleStore {
	return &HuddleStore{
		huddles: make(map[string]*Huddle),
	}
}

func (s *HuddleStore) Create(createdBy string) *Huddle {
	s.mu.Lock()
	defer s.mu.Unlock()

	huddleID := uuid.New().String()
	huddle := &Huddle{
		ID:           huddleID,
		ChannelName:  fmt.Sprintf("huddle_%s", huddleID[:8]),
		CreatedBy:    createdBy,
		CreatedAt:    time.Now().UTC(),
		Participants: []string{},
	}
	s.huddles[huddleID] = huddle
	return huddle
}

func (s *HuddleStore) GetByChannel(channelName string) (*Huddle, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, huddle := range s.huddles {
		if huddle.ChannelName == channelName {
			return huddle, true
		}
	}
	return nil, false
}

func (s *HuddleStore) Get(huddleID string) (*Huddle, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	huddle, exists := s.huddles[huddleID]
	return huddle, exists
}

func (s *HuddleStore) List() []*Huddle {
	s.mu.RLock()
	defer s.mu.RUnlock()

	huddles := make([]*Huddle, 0, len(s.huddles))
	for _, huddle := range s.huddles {
		huddles = append(huddles, huddle)
	}
	return huddles
}

func (s *HuddleStore) Join(huddleID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	huddle, exists := s.huddles[huddleID]
	if !exists {
		return fmt.Errorf("huddle not found")
	}

	// Check if already joined
	for _, participant := range huddle.Participants {
		if participant == userID {
			return nil // Already in huddle
		}
	}

	huddle.Participants = append(huddle.Participants, userID)
	return nil
}

func (s *HuddleStore) Leave(huddleID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	huddle, exists := s.huddles[huddleID]
	if !exists {
		return fmt.Errorf("huddle not found")
	}

	// Remove participant
	for i, participant := range huddle.Participants {
		if participant == userID {
			huddle.Participants = append(huddle.Participants[:i], huddle.Participants[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("user not in huddle")
}

func (s *HuddleStore) End(huddleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.huddles[huddleID]; !exists {
		return fmt.Errorf("huddle not found")
	}

	delete(s.huddles, huddleID)
	return nil
}

func (s *HuddleStore) EndByChannel(channelName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, huddle := range s.huddles {
		if huddle.ChannelName == channelName {
			delete(s.huddles, id)
			return nil
		}
	}
	return fmt.Errorf("huddle not found")
}

func (s *HuddleStore) GetOrCreate(channelName, userID string) *Huddle {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if huddle exists
	for _, huddle := range s.huddles {
		if huddle.ChannelName == channelName {
			return huddle
		}
	}

	// Create new huddle
	huddleID := uuid.New().String()
	huddle := &Huddle{
		ID:           huddleID,
		ChannelName:  channelName,
		CreatedBy:    userID,
		CreatedAt:    time.Now().UTC(),
		Participants: []string{},
	}
	s.huddles[huddleID] = huddle
	return huddle
}

func (s *HuddleStore) JoinByChannel(channelName, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, huddle := range s.huddles {
		if huddle.ChannelName == channelName {
			// Check if already joined
			for _, participant := range huddle.Participants {
				if participant == userID {
					return // Already in huddle
				}
			}
			huddle.Participants = append(huddle.Participants, userID)
			return
		}
	}
}

func main() {
	// Load .env file if present (silent fail if not found)
	_ = godotenv.Load()

	appID = os.Getenv("APP_ID")
	appCertificate = os.Getenv("APP_CERTIFICATE")

	if appID == "" || appCertificate == "" {
		log.Fatal("Error: APP_ID and APP_CERTIFICATE environment variables are required.")
	}

	// Initialize in-memory huddle store
	huddleStore = NewHuddleStore()
	log.Println("Initialized in-memory huddle store")

	api := gin.Default()

	// CORS middleware
	api.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Token endpoints with automatic huddle management
	api.GET("rtc/:channelName/:role/:tokenType/:uid", getRtcToken)
	api.GET("rtm/:uid/", getRtmToken)
	api.GET("rte/:channelName/:role/:tokenType/:uid/", getBothRokens)

	// Dedicated huddle management endpoints
	api.POST("/huddle/create", createHuddle)           // Explicitly create a huddle
	api.POST("/huddle/join", joinHuddle)               // Explicitly join a huddle
	api.POST("/huddle/leave", leaveHuddle)             // Leave a huddle
	api.POST("/huddle/end", endHuddleByID)             // End huddle by ID
	api.GET("/huddle/list", listHuddles)               // List all huddles

	// Alternative huddle query endpoints
	api.GET("/huddles", listHuddles)                   // List all active huddles
	api.DELETE("/huddles/:channelName", endHuddleByChannel) // End by channel name

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Telex Agora Huddle Backend (In-Memory) on port %s", port)
	api.Run(":" + port)
}

// Dedicated Huddle Management Handlers

func createHuddle(c *gin.Context) {
	var req struct {
		CreatedBy string `json:"created_by" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "created_by is required"})
		return
	}

	huddle := huddleStore.Create(req.CreatedBy)

	c.JSON(201, gin.H{
		"huddle_id":    huddle.ID,
		"channel_name": huddle.ChannelName,
		"created_by":   huddle.CreatedBy,
		"created_at":   huddle.CreatedAt.Format(time.RFC3339),
		"app_id":       appID,
	})
}

func joinHuddle(c *gin.Context) {
	var req struct {
		HuddleID string `json:"huddle_id" binding:"required"`
		UserID   string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "huddle_id and user_id are required"})
		return
	}

	if err := huddleStore.Join(req.HuddleID, req.UserID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Successfully joined huddle"})
}

func leaveHuddle(c *gin.Context) {
	var req struct {
		HuddleID string `json:"huddle_id" binding:"required"`
		UserID   string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "huddle_id and user_id are required"})
		return
	}

	if err := huddleStore.Leave(req.HuddleID, req.UserID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Successfully left huddle"})
}

func endHuddleByID(c *gin.Context) {
	var req struct {
		HuddleID string `json:"huddle_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "huddle_id is required"})
		return
	}

	if err := huddleStore.End(req.HuddleID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Huddle ended successfully"})
}

func endHuddleByChannel(c *gin.Context) {
	channelName := c.Param("channelName")

	if err := huddleStore.EndByChannel(channelName); err != nil {
		c.JSON(404, gin.H{"error": "Huddle not found"})
		return
	}

	c.JSON(200, gin.H{"message": "Huddle ended successfully"})
}

func listHuddles(c *gin.Context) {
	huddles := huddleStore.List()

	response := make([]gin.H, 0, len(huddles))
	for _, huddle := range huddles {
		response = append(response, gin.H{
			"huddle_id":         huddle.ID,
			"channel_name":      huddle.ChannelName,
			"created_by":        huddle.CreatedBy,
			"created_at":        huddle.CreatedAt.Format(time.RFC3339),
			"participant_count": len(huddle.Participants),
			"participants":      huddle.Participants,
		})
	}

	c.JSON(200, gin.H{"huddles": response})
}

// Unified Token Handlers with Auto Huddle Management

func getRtcToken(c *gin.Context) {
	// get param values
	channelName, tokenType, uidStr, role, expireTimestamp, err := parseRtcParams(c)
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400,
			gin.H{"message": "Error Generating RTC token: " + err.Error(),
				"status": 400,
			})
		return
	}

	// Automatically create or get huddle
	huddle := huddleStore.GetOrCreate(channelName, uidStr)

	// Track participant joining
	huddleStore.JoinByChannel(channelName, uidStr)

	// generate the token
	rtcToken, tokenErr := generateRtcToken(channelName, uidStr, tokenType, role, expireTimestamp)
	if tokenErr != nil {
		log.Println("Error generating RTC token: ", tokenErr)
		c.Error(tokenErr)
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"error":  "Error generating RTC token: " + tokenErr.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"rtcToken":    rtcToken,
			"channelName": channelName,
			"huddleId":    huddle.ID,
			"appId":       appID,
		})
	}
}

func getRtmToken(c *gin.Context) {
	// get param values
	uidStr, expireTimestamp, err := parseRtmParams(c)
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{
			"status":  400,
			"message": "Error Generating RTM token: " + err.Error(),
		})
		return
	}
	// build rtm token
	rtmToken, tokenErr := rtmtokenbuilder2.BuildToken(appID, appCertificate, uidStr, expireTimestamp, "")
	// return rtm token
	if tokenErr != nil {
		log.Println(err)
		c.Error(tokenErr)
		errMsg := "Error generating RTM token: " + tokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"error":  errMsg,
		})
		return
	}
	c.JSON(200, gin.H{
		"rtmToken": rtmToken,
	})
}

func getBothRokens(c *gin.Context) {
	// get param values
	channelName, tokenType, uidStr, role, expireTimestamp, rtcParamErr := parseRtcParams(c)
	if rtcParamErr != nil {
		c.Error(rtcParamErr)
		c.AbortWithStatusJSON(400, gin.H{
			"status":  400,
			"message": "Error Generating RTC token params: " + rtcParamErr.Error(),
		})
		return
	}

	// Automatically create or get huddle
	huddle := huddleStore.GetOrCreate(channelName, uidStr)

	// Track participant joining
	huddleStore.JoinByChannel(channelName, uidStr)

	// generate rtc token
	rtcToken, rtcTokenErr := generateRtcToken(channelName, uidStr, tokenType, role, expireTimestamp)
	// generate rtm token
	rtmToken, rtmTokenErr := rtmtokenbuilder2.BuildToken(appID, appCertificate, uidStr, expireTimestamp, "")
	// return both tokens
	if rtcTokenErr != nil {
		log.Println(rtcTokenErr)
		c.Error(rtcTokenErr)
		errMsg := "Error generating RTC token: " + rtcTokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status":  400,
			"message": errMsg,
		})
		return
	} else if rtmTokenErr != nil {
		log.Println(rtmTokenErr)
		c.Error(rtmTokenErr)
		errMsg := "Error generating RTM token: " + rtmTokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status":  400,
			"message": errMsg,
		})
		return
	} else {
		c.JSON(200, gin.H{
			"rtcToken":    rtcToken,
			"rtmToken":    rtmToken,
			"channelName": channelName,
			"huddleId":    huddle.ID,
			"appId":       appID,
		})
	}
}

func parseRtcParams(c *gin.Context) (channelName, tokenType, uidStr string, role rtctokenbuilder2.Role, expireTimestamp uint32, err error) {
	// get param values
	channelName = c.Param("channelName")
	roleStr := c.Param("role")
	tokenType = c.Param("tokenType")
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry", "3600")

	if roleStr == "publisher" {
		role = rtctokenbuilder2.RolePublisher
	} else {
		role = rtctokenbuilder2.RoleSubscriber
	}

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		err = fmt.Errorf("failed to parse expireTime: %s, causing erro: %s", expireTime, parseErr)
	}
	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeInSeconds

	return channelName, tokenType, uidStr, role, expireTimestamp, err
}

func parseRtmParams(c *gin.Context) (uidStr string, expireTimestamp uint32, err error) {
	// get param values
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry", "3600")

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		err = fmt.Errorf("failed to parse expireTime: %s, causing erro: %s", expireTime, parseErr)
	}
	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeInSeconds

	return uidStr, expireTimestamp, err
}

func generateRtcToken(channelName, uidStr, tokenType string, role rtctokenbuilder2.Role, expireTimestamp uint32) (rtcToken string, err error) {
	// check token type
	if tokenType == "userAccount" {
		rtcToken, err = rtctokenbuilder2.BuildTokenWithAccount(appID, appCertificate, channelName, uidStr, role, expireTimestamp)
		return rtcToken, err
	} else if tokenType == "uid" {
		uid64, parseErr := strconv.ParseUint(uidStr, 10, 64)
		if parseErr != nil {
			err = fmt.Errorf("Failed to parse uidStr: %s, to uint causing error: %s", uidStr, parseErr)
			return "", err
		}
		uid := uint32(uid64)
		rtcToken, err = rtctokenbuilder2.BuildTokenWithUid(appID, appCertificate, channelName, uid, role, expireTimestamp)
		return rtcToken, err
	} else {
		err = fmt.Errorf("failed to generate RTC token for unknown tokenType: %s", tokenType)
		log.Println(err)
		return "", err
	}
}
