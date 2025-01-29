package api

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/go-chi/chi/v5"
)

type timelineCtr struct {
	meta metastore.TimelineStore
}

func (c *timelineCtr) mountRouter(r chi.Router) {
	r.Get("/{traceid}", c.get)
	r.Get("/mapping", c.getMapping)
}

func (c *timelineCtr) get(w http.ResponseWriter, r *http.Request) {
	traceid := chi.URLParam(r, "traceid")

	logs, err := c.meta.Get(r.Context(), traceid, metastore.TimelineQueryOptions{})
	if err != nil {
		writeError(w, &Error{
			Code:    "error",
			Message: err.Error(),
		})

		return
	}
	writeJSON(w, logs)
}

func (c *timelineCtr) getMapping(w http.ResponseWriter, r *http.Request) {
	mapping, err := c.meta.GetMapping(r.Context())
	if err != nil {
		writeError(w, &Error{
			Code:    "error",
			Message: err.Error(),
		})

		return
	}
	writeJSON(w, mapping)
}

// {
// 	"data": {
// 	  "otel-v1-apm-service-map": {
// 		"mappings": {
// 		  "date_detection": false,
// 		  "dynamic_templates": [
// 			{
// 			  "strings_as_keyword": {
// 				"mapping": { "ignore_above": 1024, "type": "keyword" },
// 				"match_mapping_type": "string"
// 			  }
// 			}
// 		  ],
// 		  "properties": {
// 			"destination": {
// 			  "properties": {
// 				"domain": { "ignore_above": 1024, "type": "keyword" },
// 				"resource": { "ignore_above": 1024, "type": "keyword" }
// 			  }
// 			},
// 			"hashId": { "ignore_above": 1024, "type": "keyword" },
// 			"kind": { "ignore_above": 1024, "type": "keyword" },
// 			"serviceName": { "ignore_above": 1024, "type": "keyword" },
// 			"target": {
// 			  "properties": {
// 				"domain": { "ignore_above": 1024, "type": "keyword" },
// 				"resource": { "ignore_above": 1024, "type": "keyword" }
// 			  }
// 			},
// 			"traceGroupName": { "ignore_above": 1024, "type": "keyword" }
// 		  }
// 		}
// 	  }
// 	}
//   }

// {
// 	"data": {
// 	  "otel-v1-apm-span-000001": {
// 		"mappings": {
// 		  "date_detection": false,
// 		  "dynamic_templates": [
// 			{
// 			  "resource_attributes_map": {
// 				"mapping": { "type": "keyword" },
// 				"path_match": "resource.attributes.*"
// 			  }
// 			},
// 			{
// 			  "span_attributes_map": {
// 				"mapping": { "type": "keyword" },
// 				"path_match": "span.attributes.*"
// 			  }
// 			}
// 		  ],
// 		  "properties": {
// 			"droppedAttributesCount": { "type": "long" },
// 			"droppedEventsCount": { "type": "long" },
// 			"droppedLinksCount": { "type": "long" },
// 			"durationInNanos": { "type": "long" },
// 			"endTime": { "type": "date_nanos" },
// 			"events": {
// 			  "properties": {
// 				"attributes": { "type": "object" },
// 				"droppedAttributesCount": { "type": "long" },
// 				"name": {
// 				  "fields": {
// 					"keyword": { "ignore_above": 256, "type": "keyword" }
// 				  },
// 				  "type": "text"
// 				},
// 				"time": { "type": "date_nanos" }
// 			  },
// 			  "type": "nested"
// 			},
// 			"instrumentationScope": {
// 			  "properties": {
// 				"name": {
// 				  "fields": {
// 					"keyword": { "ignore_above": 256, "type": "keyword" }
// 				  },
// 				  "type": "text"
// 				}
// 			  }
// 			},
// 			"kind": { "ignore_above": 128, "type": "keyword" },
// 			"links": { "type": "nested" },
// 			"name": { "ignore_above": 1024, "type": "keyword" },
// 			"parentSpanId": { "ignore_above": 256, "type": "keyword" },
// 			"resource": {
// 			  "properties": {
// 				"attributes": {
// 				  "properties": { "service@name": { "type": "keyword" } }
// 				}
// 			  }
// 			},
// 			"serviceName": { "type": "keyword" },
// 			"span": {
// 			  "properties": {
// 				"attributes": {
// 				  "properties": {
// 					"action": { "type": "keyword" },
// 					"api@version": { "type": "keyword" },
// 					"callpath": { "type": "keyword" },
// 					"http@method": { "type": "keyword" },
// 					"http@route": { "type": "keyword" },
// 					"instance": { "type": "keyword" },
// 					"instance@manager": { "type": "keyword" },
// 					"invoker": { "type": "keyword" },
// 					"namespace": { "type": "keyword" },
// 					"state": { "type": "keyword" },
// 					"status": { "type": "keyword" },
// 					"track": { "type": "keyword" },
// 					"workflow": { "type": "keyword" }
// 				  }
// 				}
// 			  }
// 			},
// 			"spanId": { "ignore_above": 256, "type": "keyword" },
// 			"startTime": { "type": "date_nanos" },
// 			"status": {
// 			  "properties": {
// 				"code": { "type": "integer" },
// 				"message": { "type": "keyword" }
// 			  }
// 			},
// 			"traceGroup": { "ignore_above": 1024, "type": "keyword" },
// 			"traceGroupFields": {
// 			  "properties": {
// 				"durationInNanos": { "type": "long" },
// 				"endTime": { "type": "date_nanos" },
// 				"statusCode": { "type": "integer" }
// 			  }
// 			},
// 			"traceId": { "ignore_above": 256, "type": "keyword" },
// 			"traceState": {
// 			  "fields": { "keyword": { "ignore_above": 256, "type": "keyword" } },
// 			  "type": "text"
// 			}
// 		  }
// 		}
// 	  }
// 	}
// }
