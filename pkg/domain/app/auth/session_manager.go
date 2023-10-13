package auth

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/datatypes"
	"sync"
	"time"
)

var errInvalidUuid = errors.New("invalid uuid")
var errAddExisting = errors.New("cannot add existing uuid again, terminating all sessions for this subject")

// sessionManager stores and manages session data
type sessionManager struct {
	sessions *datatypes.SafeMap[sessionClaims] // map[uuid]sessionClaims
}

// newSessionManager creates a new sessionManager
func newSessionManager() *sessionManager {
	sm := &sessionManager{
		sessions: datatypes.NewSafeMap[sessionClaims](),
	}

	return sm
}

// new adds the session to the map of open sessions. If session already exists
// an error is returned and all sessions for the specific subject (user) are
// removed.
func (sm *sessionManager) new(session sessionClaims) error {
	// parse uuid to a sane string for map key
	uuidString := session.UUID.String()
	if uuidString == "" {
		sm.closeSubject(session)
		return errInvalidUuid
	}

	// check if session already exists
	exists, _ := sm.sessions.Add(uuidString, session)
	if exists {
		sm.closeSubject(session)
		return errAddExisting
	}

	return nil
}

// close removes the session from the map of open sessions. If session
// doesn't exist an error is returned and all sessions for the specific
// subject (user) are removed.
func (sm *sessionManager) close(session sessionClaims) error {
	// parse uuid to a sane string for map key
	uuidString := session.UUID.String()
	if uuidString == "" {
		sm.closeSubject(session)
		return errInvalidUuid
	}

	// remove and check if trying to remove non-existent
	err := sm.sessions.DeleteKey(uuidString)
	if err != nil {
		sm.closeSubject(session)
		return err
	}

	return nil
}

// refresh confirms the oldSession is present, removes it, and then adds the new
// session in its place. If the session doesn't exist or the new session
// already exists an error is returned and all sessions for the specific subject
// (user) are removed.
func (sm *sessionManager) refresh(oldSession, newSession sessionClaims) error {
	// remove old session (error if doesn't exist, so this is validation)
	err := sm.close(oldSession)

	if err != nil {
		// closeSubject already called by sm.close()
		return err
	}

	// add new session (error if already exists)
	err = sm.new(newSession)
	if err != nil {
		// closeSubject already called by sm.close()
		return err
	}

	return nil
}

// closeSubject deletes all sessions where the session's Subject is equal to
// the specified sessionClaims' Subject
func (sm *sessionManager) closeSubject(sc sessionClaims) {
	// delete func for close subject
	deleteFunc := func(k string, v sessionClaims) bool {
		// if map value Subject == this func param's (sc's) Subject, return true
		return v.Subject == sc.Subject
	}

	// run func against sessions map
	sm.sessions.DeleteFunc(deleteFunc)
}

// startCleanerService starts a goroutine that is an indefinite for loop
// that checks for expired sessions and removes them. This is to
// prevent the accumulation of expired sessions that were never
// formally logged out of.
func (service *Service) startCleanerService(ctx context.Context, wg *sync.WaitGroup) {
	// log start and update wg
	service.logger.Info("starting auth session cleaner service")
	wg.Add(1)

	// delete func that checks values for expired session
	deleteFunc := func(k string, v sessionClaims) bool {
		if v.ExpiresAt.Unix() <= time.Now().Unix() {
			// if expiration has passed, delete
			service.logger.Infof("user '%s' logged out (expired)", v.Subject)
			return true
		}

		// else don't delete (valid)
		return false
	}

	go func() {
		// wait time is based on expiration of sessions (refresh)
		waitTime := 2 * refreshTokenExpiration
		for {
			select {
			case <-ctx.Done():
				// exit
				service.logger.Info("auth session cleaner service shutdown complete")
				wg.Done()
				return

			case <-time.After(waitTime):
				// continue and run
			}

			// run delete func against sessions map
			service.sessionManager.sessions.DeleteFunc(deleteFunc)
		}
	}()
}
