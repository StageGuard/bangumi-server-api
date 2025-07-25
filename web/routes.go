// SPDX-License-Identifier: AGPL-3.0-only
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>

package web

import (
	"github.com/labstack/echo/v4"
	"github.com/redis/rueidis"

	"github.com/bangumi/server/web/handler"
	"github.com/bangumi/server/web/handler/character"
	"github.com/bangumi/server/web/handler/common"
	"github.com/bangumi/server/web/handler/index"
	"github.com/bangumi/server/web/handler/person"
	"github.com/bangumi/server/web/handler/subject"
	"github.com/bangumi/server/web/handler/user"
	"github.com/bangumi/server/web/mw"
	"github.com/bangumi/server/web/mw/ua"
	"github.com/bangumi/server/web/req"
)

// AddRouters add all router and default 404 Handler to app.
//
//nolint:funlen
func AddRouters(
	app *echo.Echo,
	common common.Common,
	rueidis rueidis.Client,
	h handler.Handler,
	userHandler user.User,
	personHandler person.Person,
	characterHandler character.Character,
	subjectHandler subject.Subject,
	indexHandler index.Handler,
) {
	app.GET("/", indexPage())

	app.Use(ua.DisableDefaultHTTPLibrary)
	app.Use(ua.DisableBrokenUA)

	v0 := app.Group("/v0", common.MiddlewareAccessTokenAuth)
	v0.Use(mw.RateLimit(common.Config, rueidis))

	v0.POST("/search/subjects", h.SearchSubjects)
	v0.POST("/search/characters", h.SearchCharacters)
	v0.POST("/search/persons", h.SearchPersons)

	subjectHandler.Routes(v0)

	v0.GET("/persons/:id", personHandler.Get)
	v0.GET("/persons/:id/image", personHandler.GetImage)
	v0.GET("/persons/:id/subjects", personHandler.GetRelatedSubjects)
	v0.GET("/persons/:id/characters", personHandler.GetRelatedCharacters)
	v0.POST("/persons/:id/collect", personHandler.CollectPerson, mw.NeedLogin)
	// TODO: wait for soft delete
	// v0.DELETE("/persons/:id/collect", personHandler.UncollectPerson, mw.NeedLogin)

	v0.GET("/characters/:id", characterHandler.Get)
	v0.GET("/characters/:id/image", characterHandler.GetImage)
	v0.GET("/characters/:id/subjects", characterHandler.GetRelatedSubjects)
	v0.GET("/characters/:id/persons", characterHandler.GetRelatedPersons)
	v0.POST("/characters/:id/collect", characterHandler.CollectCharacter, mw.NeedLogin)
	// TODO: wait for soft delete
	// v0.DELETE("/characters/:id/collect", characterHandler.UncollectCharacter, mw.NeedLogin)

	v0.GET("/episodes/:id", h.GetEpisode)
	v0.GET("/episodes", h.ListEpisode)

	// echo 中间件从前往后运行按顺序
	v0.GET("/me", userHandler.GetCurrent)
	v0.GET("/users/:username", userHandler.Get)
	v0.GET("/users/:username/avatar", userHandler.GetAvatar)
	v0.GET("/users/:username/collections", userHandler.ListSubjectCollection)
	v0.GET("/users/:username/collections/:subject_id", userHandler.GetSubjectCollection)

	v0.GET("/users/-/collections/-/episodes/:episode_id", userHandler.GetEpisodeCollection, mw.NeedLogin)
	v0.PUT("/users/-/collections/-/episodes/:episode_id", userHandler.PutEpisodeCollection, req.JSON, mw.NeedLogin)
	v0.GET("/users/-/collections/:subject_id/episodes", userHandler.GetSubjectEpisodeCollection, mw.NeedLogin)
	v0.PATCH("/users/-/collections/:subject_id", userHandler.PatchSubjectCollection, req.JSON, mw.NeedLogin)
	v0.POST("/users/-/collections/:subject_id", userHandler.PostSubjectCollection, req.JSON, mw.NeedLogin)
	v0.PATCH("/users/-/collections/:subject_id/episodes",
		userHandler.PatchEpisodeCollectionBatch, req.JSON, mw.NeedLogin)

	v0.GET("/users/:username/collections/-/characters", userHandler.ListCharacterCollection)
	v0.GET("/users/:username/collections/-/characters/:character_id", userHandler.GetCharacterCollection)
	v0.GET("/users/:username/collections/-/persons", userHandler.ListPersonCollection)
	v0.GET("/users/:username/collections/-/persons/:person_id", userHandler.GetPersonCollection)

	{
		i := indexHandler
		v0.GET("/indices/:id", i.GetIndex)
		v0.GET("/indices/:id/subjects", i.GetIndexSubjects)
		// indices
		v0.POST("/indices", i.NewIndex, req.JSON, mw.NeedLogin)
		v0.PUT("/indices/:id", i.UpdateIndex, req.JSON, mw.NeedLogin)
		// indices subjects
		v0.POST("/indices/:id/subjects", i.AddIndexSubject, req.JSON, mw.NeedLogin)
		v0.PUT("/indices/:id/subjects/:subject_id", i.UpdateIndexSubject, req.JSON, mw.NeedLogin)
		v0.DELETE("/indices/:id/subjects/:subject_id", i.RemoveIndexSubject, mw.NeedLogin)
		// collect
		v0.POST("/indices/:id/collect", i.CollectIndex, mw.NeedLogin)
		v0.DELETE("/indices/:id/collect", i.UncollectIndex, mw.NeedLogin)
	}

	v0.GET("/revisions/persons/:id", h.GetPersonRevision)
	v0.GET("/revisions/persons", h.ListPersonRevision)
	v0.GET("/revisions/subjects/:id", h.GetSubjectRevision)
	v0.GET("/revisions/subjects", h.ListSubjectRevision)
	v0.GET("/revisions/characters/:id", h.GetCharacterRevision)
	v0.GET("/revisions/characters", h.ListCharacterRevision)

	v0.GET("/revisions/episodes/:id", h.GetEpisodeRevision)
	v0.GET("/revisions/episodes", h.ListEpisodeRevision)
	v0.Any("/*", globalNotFoundHandler)

	// default 404 Handler, all router should be added before this router
	app.Any("/*", globalNotFoundHandler)
}
