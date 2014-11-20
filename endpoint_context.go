package elasticthought

import (
	"errors"
	"fmt"

	"github.com/couchbaselabs/cbfs/client"
	"github.com/couchbaselabs/logg"
	"github.com/gin-gonic/gin"
	"github.com/tleyden/go-couch"
)

type EndpointContext struct {
	Configuration Configuration
}

func (e EndpointContext) CreateSolverEndpoint(c *gin.Context) {

	user := c.MustGet(MIDDLEWARE_KEY_USER).(User)
	db := c.MustGet(MIDDLEWARE_KEY_DB).(couch.Database)
	logg.LogTo("REST", "user: %v db: %v", user, db)

	solver := NewSolver()

	// bind the input struct to the JSON request
	if ok := c.Bind(solver); !ok {
		errMsg := fmt.Sprintf("Invalid input")
		c.Fail(400, errors.New(errMsg))
		return
	}

	logg.LogTo("REST", "solver: %+v", solver)

	// save solver in db
	solver, err := solver.Insert(db)
	if err != nil {
		c.Fail(500, err)
		return
	}

	// Create a cbfs client
	cbfs, err := cbfsclient.New(e.Configuration.CbfsUrl)
	if err != nil {
		errMsg := fmt.Errorf("Error creating cbfs client: %v", err)
		c.Fail(500, errMsg)
		return
	}
	logg.LogTo("REST", "cbfs: %+v", cbfs)

	// download contents of solver-spec-url into cbfs://<solver-id>/spec.prototxt
	// and update solver object's solver-spec-url with cbfs url

	solver, err = solver.SaveSpec(db)
	if err != nil {
		c.Fail(500, err)
		return
	}

	// return solver object
	c.JSON(201, *solver)

}
