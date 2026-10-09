package server

import (
	"testing"

	"github.com/stretchr/testify/require"
	"heckel.io/ntfy/v2/user"
	"heckel.io/ntfy/v2/util"
)

func TestServer_TopicsGet_Unauthorized(t *testing.T) {
	forEachBackend(t, func(t *testing.T, databaseURL string) {
		s := newTestServer(t, newTestConfigWithAuthFile(t, databaseURL))

		rr := request(t, s, "GET", "/v1/topics", "", nil)
		require.Equal(t, 401, rr.Code)
		require.Equal(t, 40101, toHTTPError(t, rr.Body.String()).Code)
	})
}

func TestServer_TopicsGet_InvalidCredentials(t *testing.T) {
	forEachBackend(t, func(t *testing.T, databaseURL string) {
		s := newTestServer(t, newTestConfigWithAuthFile(t, databaseURL))
		require.Nil(t, s.userManager.AddUser("ben", "ben", user.RoleUser, false))

		rr := request(t, s, "GET", "/v1/topics", "", map[string]string{
			"Authorization": util.BasicAuth("ben", "wrong"),
		})
		require.Equal(t, 401, rr.Code)
	})
}

func TestServer_TopicsGet_AuthDisabled(t *testing.T) {
	s := newTestServer(t, newTestConfig(t, "")) // No auth-file, so there is no user manager

	rr := request(t, s, "GET", "/v1/topics", "", nil)
	require.Equal(t, 404, rr.Code)
}

func TestServer_TopicsGet_Empty(t *testing.T) {
	forEachBackend(t, func(t *testing.T, databaseURL string) {
		s := newTestServer(t, newTestConfigWithAuthFile(t, databaseURL))
		require.Nil(t, s.userManager.AddUser("ben", "ben", user.RoleUser, false))

		rr := request(t, s, "GET", "/v1/topics", "", map[string]string{
			"Authorization": util.BasicAuth("ben", "ben"),
		})
		require.Equal(t, 200, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		require.JSONEq(t, `{"topics":[]}`, rr.Body.String())
	})
}

func TestServer_TopicsGet_UserSeesReadableTopicsOnly(t *testing.T) {
	forEachBackend(t, func(t *testing.T, databaseURL string) {
		c := newTestConfigWithAuthFile(t, databaseURL)
		c.AuthDefault = user.PermissionDenyAll
		s := newTestServer(t, c)
		require.Nil(t, s.userManager.AddUser("ben", "ben", user.RoleUser, false))
		require.Nil(t, s.userManager.AllowAccess("ben", "mytopic", user.PermissionRead))
		require.Nil(t, s.userManager.AllowAccess("ben", "writeonly", user.PermissionWrite))
		require.Nil(t, s.messageCache.AddMessage(newMessageWithTimestamp("mytopic", "message 1", 1700000100)))
		require.Nil(t, s.messageCache.AddMessage(newMessageWithTimestamp("writeonly", "message 2", 1700000200)))
		require.Nil(t, s.messageCache.AddMessage(newMessageWithTimestamp("othertopic", "message 3", 1700000300)))

		rr := request(t, s, "GET", "/v1/topics", "", map[string]string{
			"Authorization": util.BasicAuth("ben", "ben"),
		})
		require.Equal(t, 200, rr.Code)
		require.JSONEq(t, `{"topics":[{"topic":"mytopic","last_time":1700000100}]}`, rr.Body.String())
	})
}

func TestServer_TopicsGet_UserSeesTopicsReadableByEveryone(t *testing.T) {
	forEachBackend(t, func(t *testing.T, databaseURL string) {
		c := newTestConfigWithAuthFile(t, databaseURL)
		c.AuthDefault = user.PermissionDenyAll
		s := newTestServer(t, c)
		require.Nil(t, s.userManager.AddUser("ben", "ben", user.RoleUser, false))
		require.Nil(t, s.userManager.AllowAccess(user.Everyone, "announcements", user.PermissionRead))
		require.Nil(t, s.messageCache.AddMessage(newMessageWithTimestamp("announcements", "message 1", 1700000100)))
		require.Nil(t, s.messageCache.AddMessage(newMessageWithTimestamp("othertopic", "message 2", 1700000200)))

		rr := request(t, s, "GET", "/v1/topics", "", map[string]string{
			"Authorization": util.BasicAuth("ben", "ben"),
		})
		require.Equal(t, 200, rr.Code)
		require.JSONEq(t, `{"topics":[{"topic":"announcements","last_time":1700000100}]}`, rr.Body.String())
	})
}

func TestServer_TopicsGet_AdminSeesAllTopics(t *testing.T) {
	forEachBackend(t, func(t *testing.T, databaseURL string) {
		c := newTestConfigWithAuthFile(t, databaseURL)
		c.AuthDefault = user.PermissionDenyAll
		s := newTestServer(t, c)
		require.Nil(t, s.userManager.AddUser("phil", "phil", user.RoleAdmin, false))
		require.Nil(t, s.messageCache.AddMessage(newMessageWithTimestamp("mytopic", "message 1", 1700000100)))
		require.Nil(t, s.messageCache.AddMessage(newMessageWithTimestamp("othertopic", "message 2", 1700000200)))

		rr := request(t, s, "GET", "/v1/topics", "", map[string]string{
			"Authorization": util.BasicAuth("phil", "phil"),
		})
		require.Equal(t, 200, rr.Code)
		require.JSONEq(t, `{"topics":[
			{"topic":"othertopic","last_time":1700000200},
			{"topic":"mytopic","last_time":1700000100}
		]}`, rr.Body.String())
	})
}

func TestServer_TopicsGet_SortedByLastTimeDescending(t *testing.T) {
	forEachBackend(t, func(t *testing.T, databaseURL string) {
		s := newTestServer(t, newTestConfigWithAuthFile(t, databaseURL))
		require.Nil(t, s.userManager.AddUser("phil", "phil", user.RoleAdmin, false))
		require.Nil(t, s.messageCache.AddMessage(newMessageWithTimestamp("mytopic", "message 1", 1700000100)))
		require.Nil(t, s.messageCache.AddMessage(newMessageWithTimestamp("othertopic", "message 2", 1700000300)))
		require.Nil(t, s.messageCache.AddMessage(newMessageWithTimestamp("thirdtopic", "message 3", 1700000200)))

		rr := request(t, s, "GET", "/v1/topics", "", map[string]string{
			"Authorization": util.BasicAuth("phil", "phil"),
		})
		require.Equal(t, 200, rr.Code)
		require.JSONEq(t, `{"topics":[
			{"topic":"othertopic","last_time":1700000300},
			{"topic":"thirdtopic","last_time":1700000200},
			{"topic":"mytopic","last_time":1700000100}
		]}`, rr.Body.String())
	})
}

func TestServer_TopicsGet_CacheDisabled(t *testing.T) {
	forEachBackend(t, func(t *testing.T, databaseURL string) {
		c := newTestConfigWithAuthFile(t, databaseURL)
		c.CacheDuration = 0 // Messages are not cached at all
		s := newTestServer(t, c)
		require.Nil(t, s.userManager.AddUser("phil", "phil", user.RoleAdmin, false))

		rr := request(t, s, "PUT", "/mytopic", "message", map[string]string{
			"Authorization": util.BasicAuth("phil", "phil"),
		})
		require.Equal(t, 200, rr.Code)

		rr = request(t, s, "GET", "/v1/topics", "", map[string]string{
			"Authorization": util.BasicAuth("phil", "phil"),
		})
		require.Equal(t, 200, rr.Code)
		require.JSONEq(t, `{"topics":[]}`, rr.Body.String())
	})
}
