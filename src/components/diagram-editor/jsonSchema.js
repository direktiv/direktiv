// COMMON

const CommonSchemaDefinitionConsumeEvent = {
    "type": "object",
    "title": "Event Definition",
    "description": "Event to consume.",
    "required": [
        "type"
    ],
    "properties": {
        "type": {
            "type": "string",
            "title": "Type",
            "description": "CloudEvent type."
        },
        "context": {
            "type": "object",
            "title": "Context",
            "description": "Key value pairs for CloudEvent context values that must match.",
            "additionalProperties": {
                "type": "string"
            },
        }
    }
}

const CommonSchemaDefinitionFiles = {
    "type": "array",
    "minItems": 0,
    "title": "Files",
    "description": "Action file definitions.",
    "items": {
        "type": "object",
        "required": [
            "key"
        ],
        "properties": {
            "key": {
                "type": "string",
                "title": "Key",
                "description": "Key used to select variable."
            },
            "scope": {
                "title": "Scope",
                "description": "Scope used to select variable. Defaults to 'instance', but can be 'workflow' or 'namespace'.",
                "type": "string",
                "enum": [
                    "instance",
                    "workflow",
                    "namespace"
                ],
                "default": "workflow"
            },
            "as": {
                "title": "As",
                "description": "Set the filename of the file. The default is the same as the key.",
                "type": "string"
            },
            "type": {
                "title": "Type",
                "description": "How to treat the file. Options include 'plain', 'base64', 'tar', 'tar.gz'.",
                "type": "string"
            }
        }
    }
}

const CommonSchemaDefinitionTimeout = {
    "type": "string",
    "title": "Timeout",
    "description": "Duration to wait for action to complete (ISO8601)."
}

export const CommonSchemaDefinitionStateFields = {
    "transform": {
        "title": "Transform",
        "description": "jq command to transform the state's data output.",
        "type": "object",
        "properties": {
            "selectionType": {
                "enum": [
                    "JQ Query",
                    "Key Value",
                    "YAML",
                    "Javascript"
                ],
                "default": "JQ Query"
            }
        },
        "allOf": [
            {
                "if": {
                    "properties": {
                        "selectionType": {
                            "const": "JQ Query"
                        }
                    }
                },
                "then": {
                    "properties": {
                        "jqQuery": {
                            "type": "string"
                        }
                    }
                }
            },
            {
                "if": {
                    "properties": {
                        "selectionType": {
                            "const": "YAML"
                        }
                    }
                },
                "then": {
                    "properties": {
                        "rawYAML": {
                            "title": "YAML",
                            "type": "string",
                            "description": "Raw YAML object representation of data.",
                        }
                    }
                }
            },
            {
                "if": {
                    "properties": {
                        "selectionType": {
                            "const": "Javascript"
                        }
                    }
                },
                "then": {
                    "properties": {
                        "js": {
                            "title": "Javascript",
                            "type": "string",
                            "description": "TODO: Javascript",
                        }
                    }
                }
            },
            {
                "if": {
                    "properties": {
                        "selectionType": {
                            "const": "Key Value"
                        }
                    }
                },
                "then": {
                    "properties": {
                        "keyValue": {
                            "type": "object",
                            "title": "Key Value",
                            "description": "Key Values representation of data.",
                            "additionalProperties": {
                                "type": "string"
                            },
                        }
                    }
                }
            }
        ]
    },
    "log": {
        "type": "string",
        "title": "Log",
        "description": "jq command to generate data for instance-logging."
    }
}


const CommonSchemaDefinitionRetry = {
    "type": "array",
    "title": "Retry Definition",
    "description": "Retry policy.",
    "maxItems": 1,
    "items": {
        "required": [
            "max_attempts",
            "codes"
        ],
        "properties": {
            "max_attempts": {
                "type": "number",
                "title": "Max Attempts",
                "description": "Maximum number of retry attempts."
            },
            "delay": {
                "type": "string",
                "title": "Delay",
                "description": "Time delay between retry attempts (ISO8601)."
            },
            "multiplier": {
                "type": "number",
                "title": "Multiplier",
                "description": "Value by which the delay is multiplied after each attempt."
            },
            "codes": {
                "type": "array",
                "title": "Codes",
                "minItems": 1,
                "description": "Regex patterns to specify which error codes to catch.",
                "items": {
                    "type": "string"
                }
            }
        }
    }
}

const CommonSchemaDefinitionAction = {
    "type": "object",
    "title": "Action Definition",
    "description": "Action to perform.",
    "required": [
        "function",
    ],
    "properties": {
        "function": {
            "enum": [
            ],
            "type": "string",
            "title": "Function",
            "description": "Name of the referenced function.",
        },
        "input": {
            ...CommonSchemaDefinitionStateFields.transform,
            "title": "Input",
            "description": "jq command to generate the input for the action."
        },
        "secrets": {
            "type": "array",
            "title": "Secrets",
            "description": "List of secrets to temporarily add to the state data under .secrets before running the input jq command.",
            "items": {
                "type": "string"
            }
        },
        "files": CommonSchemaDefinitionFiles,
        "retries": CommonSchemaDefinitionRetry
    }
}

// States

export const StateSchemaNoop = {
    "type": "object",
    "properties": {
        ...CommonSchemaDefinitionStateFields,
    }
}

export const StateSchemaConsumeEvent = {
    "type": "object",
    "required": [
        "event",
    ],
    "properties": {
        event: CommonSchemaDefinitionConsumeEvent,
        timeout: CommonSchemaDefinitionTimeout,
        ...CommonSchemaDefinitionStateFields,
    }
}

export const StateSchemaDelay = {
    "type": "object",
    "required": [
        "duration",
    ],
    "properties": {
        "duration": {
            "type": "string",
            "title": "Duration",
            "description": CommonSchemaDefinitionTimeout.description
        },
        ...CommonSchemaDefinitionStateFields,
    }
}

export const StateSchemaError = {
    "type": "object",
    "required": [
        "error",
        "message"
    ],
    "properties": {
        "error": {
            "type": "string",
            "title": "Error",
            "description": "Error code, catchable on a calling workflow.",
        },
        "message": {
            "type": "string",
            "title": "Message",
            "description": "Format string to provide more context to the error.",
        },
        "args": {
            "type": "string",
            "title": "Arguments",
            "description": "A list of jq commands to generate arguments for substitution in the message format string.",
        },
        ...CommonSchemaDefinitionStateFields,
    }
}

export const StateSchemaEventsAnd = {
    "type": "object",
    "required": [
        "events",
    ],
    "properties": {
        "events": {
            "type": "array",
            "minItems": 1,
            "title": "Events",
            "description": "Events to consume.",
            "items": {
                ...CommonSchemaDefinitionConsumeEvent,
            }
        },
        ...CommonSchemaDefinitionStateFields,
    }
}

export const StateSchemaEventXor = {
    "type": "object",
    "required": [
        "events",
    ],
    "properties": {
        "events": {
            "type": "array",
            "minItems": 1,
            "title": "Events",
            "description": "Events to consume, and what to do based on which event was received.",
            "items": {
                "type": "object",
                "required": [
                    "event"
                ],
                "properties": {
                    "event": CommonSchemaDefinitionConsumeEvent,
                    "transform": CommonSchemaDefinitionStateFields.transform,
                }
            }
        },
        "log": CommonSchemaDefinitionStateFields.log,
    }
}

export const StateSchemaForeach = {
    "type": "object",
    "required": [
        "action",
    ],
    "properties": {
        "array": {
            "type": "string",
            "title": "Array",
            "description": "jq command to produce an array of objects to loop through.",
        },
        "action": CommonSchemaDefinitionAction,
        "timeout": CommonSchemaDefinitionTimeout,
        ...CommonSchemaDefinitionStateFields,
    }
}

export const StateSchemaParallel = {
    "type": "object",
    "required": [
        "actions",
    ],
    "properties": {
        "mode": {
            "title": "Mode",
            "description": "Option types on how to complete branch execution",
            "default": "and",
            "enum": [
                "and",
                "or"
            ],
        },
        "actions": {
            "type": "array",
            "description": "List of actions to perform.",
            "title": "Actions",
            "items": {
                ...CommonSchemaDefinitionAction,
            }
        },
        "timeout": CommonSchemaDefinitionTimeout,
        ...CommonSchemaDefinitionStateFields,
    }
}

export const StateSchemaGenerateEvent = {
    "type": "object",
    "required": [
        "event",
    ],
    "properties": {
        "event": {
            "type": "object",
            "title": "Event Definition",
            "description": "Event to generate.",
            "required": [
                "type",
                "source"
            ],
            "properties": {
                "type": {
                    "type": "string",
                    "title": "Type",
                    "description": "CloudEvent type."
                },
                "source": {
                    "type": "string",
                    "title": "Source",
                    "description": "CloudEvent source."
                },
                "datacontenttype": {
                    "type": "string",
                    "title": "Data Content Type",
                    "description": "An RFC 2046 string specifying the payload content type."
                },
                "data": {
                    ...CommonSchemaDefinitionStateFields.transform,
                    "title": "Data",
                    "description": "Data to generate (payload) for the produced event."
                },
                "context": {
                    "type": "object",
                    "title": "Context",
                    "description": "Key value pairs for CloudEvent context values that must match.",
                    "additionalProperties": {
                        "type": "string"
                    },
                }
            }
        },
        ...CommonSchemaDefinitionStateFields
    }
}

export const StateSchemaGetter = {
    "type": "object",
    "required": [
        "variables"
    ],
    "properties": {
        "variables": {
            "type": "array",
            "title": "Variables",
            "description": "Variables to fetch.",
            "items": {
                "type": "object",
                "required": [
                    "key",
                    "scope"
                ],
                "properties": {
                    "key": {
                        "type": "string",
                        "title": "Key",
                        "description": "Variable name."
                    },
                    "scope": {
                        "title": "Scope",
                        "description": "Variable scope",
                        "enum": [
                            "workflow",
                            "instance",
                            "namespace"
                        ],
                        "default": "workflow"
                    },
                }
            }
        },
        ...CommonSchemaDefinitionStateFields,
    }
}

export const StateSchemaSetter = {
    "type": "object",
    "required": [
        "variables"
    ],
    "properties": {
        "variables": {
            "type": "array",
            "title": "Variables",
            "description": "Variables to push.",
            "items": {
                "type": "object",
                "required": [
                    "key",
                    "scope",
                    "value"
                ],
                "properties": {
                    "key": {
                        "type": "string",
                        "title": "Key",
                        "description": "Variable name."
                    },
                    "scope": {
                        "title": "Scope",
                        "description": "Variable scope",
                        "enum": [
                            "workflow",
                            "instance",
                            "namespace"
                        ],
                        "default": "workflow"
                    },
                    "value": {
                        ...CommonSchemaDefinitionStateFields.transform,
                        "title": "Value",
                        "description": "Value to generate variable value."
                    },
                    "mimeType": {
                        "type": "string",
                        "title": "Mime Type",
                        "description": "MimeType to store variable value as."
                    },
                }
            }
        },
        ...CommonSchemaDefinitionStateFields
    }
}

export const StateSchemaValidate = {
    "type": "object",
    "required": [
        "schema"
    ],
    "properties": {
        "subject": {
            "type": "string",
            "title": "Subject",
            "description": "jq command to select the subject of the schema validation. Defaults to '.' if unspecified."
        },
        "schema": {
            "type": "string",
            "title": "Schema",
            "description": "Name of the referenced state data schema."
        },
        ...CommonSchemaDefinitionStateFields,
    }
}


export const StateSchemaAction = {
    "type": "object",
    "properties": {
        "action": CommonSchemaDefinitionAction,
        "async": {
            "title": "Async",
            "description": "If workflow execution can continue without waiting for the action to return.",
            "type": "boolean"
        },
        "timeout": CommonSchemaDefinitionTimeout,
        ...CommonSchemaDefinitionStateFields,
    }
}

export const StateSchemaSwitch = {
    "type": "object",
    "required": [
        "conditions"
    ],
    "properties": {
        "conditions": {
            "type": "array",
            "minItems": 1,
            "title": "Conditions",
            "description": "Conditions to evaluate and determine which state to transition to next.",
            "items": {
                "type": "object",
                "required": [
                    "condition"
                ],
                "properties": {
                    "condition": {
                        "type": "string",
                        "title": "Condition",
                        "description": "jq command evaluated against state data. True if results are not empty."
                    },
                    "transform": CommonSchemaDefinitionStateFields.transform
                }
            }
        },
        "defaultTransform": {
            ...CommonSchemaDefinitionStateFields.transform,
            "title": "Default Transform",
            "descrtiption": "jq command to transform the state's data output."
        }
    }
}

// Special
const SpecialSchemaError = {
    "type": "array",
    "title": "Error Handling",
    "description": "Thrown erros will be compared against each Error in order until it finds a match.",
    "items": {
        "required": [
            "error"
        ],
        "properties": {
            "error": {
                "type": "string",
                "title": "Error",
                "description": "A glob pattern to test error codes for a match."
            }
        }
    }
}

export const SpecialSchemaScheduledStart = {
    "type": "object",
    "required": [

    ],
    "properties": {
        "cron": {
            "type": "string",
            "title": "Cron Expression",
            "description": "Cron expression to schedule workflow."
        }
    }
}

export const SpecialSchemaEventStart = {
    "type": "object",
    "required": [
        "event"
    ],
    "properties": {
        "event": {
            "type": "object",
            "title": "Event",
            "description": "Event to listen for, which can trigger the workflow.",
            "required": [
                "type"
            ],
            "properties": {
                "type": {
                    "type": "string",
                    "title": "Cloud Event",
                    "description": "CloudEvent type."
                },
                "context": {
                    "type": "object",
                    "title": "Context",
                    "description": "Key value pairs for CloudEvent context values that must match.",
                    "additionalProperties": {
                        "type": "string"
                    },
                }
            }
        }
    }
}

export const SpecialSchemaEventsXorStart = {
    "type": "object",
    "required": [
        "events"
    ],
    "properties": {
        "events": {
            "type": "array",
            "title": "Events",
            "description": "Events to listen for, which can trigger the workflow.",
            "items": {
                "type": "object",
                "required": [
                    "type"
                ],
                "properties": {
                    "type": {
                        "type": "string",
                        "title": "Cloud Event",
                        "description": "CloudEvent type."
                    },
                    "context": {
                        "type": "object",
                        "title": "Context",
                        "description": "Key value pairs for CloudEvent context values that must match.",
                        "additionalProperties": {
                            "type": "string"
                        },
                    }
                }
            }
        }
    }
}

export const SpecialSchemaEventsAndStart = {
    "type": "object",
    "required": [
        "events"
    ],
    "properties": {
        "events": {
            "type": "array",
            "title": "Events",
            "description": "Events to listen for, which can trigger the workflow.",
            "items": {
                "type": "object",
                "required": [
                    "type"
                ],
                "properties": {
                    "type": {
                        "type": "string",
                        "title": "Cloud Event",
                        "description": "CloudEvent type."
                    },
                    "context": {
                        "type": "object",
                        "title": "Context",
                        "description": "Key value pairs for CloudEvent context values that must match.",
                        "additionalProperties": {
                            "type": "string"
                        },
                    }
                }
            }
        },
        "lifespan": {
            "type": "string",
            "title": "Lifespan",
            "description": "Maximum duration an event can be stored before being discarded while waiting for other events (ISO8601).",
        },
        "correlate": {
            "type": "array",
            "title": "Correlate",
            "description": "Context keys that must exist on every event and have matching values to be grouped together.",
            "items": {
                "type": "string"
            }
        },
    }
}

export const SpecialSchemaDefaultStart = {
    "type": "object",
    "required": [

    ],
    "properties": {

    }
}

const SpecialSchemaStart = {
    "type": "object",
    "required": [
        "type"
    ],
    "properties": {
        "type": {
            "enum": [
                "default",
                "scheduled",
                "event",
                "eventsAnd",
                "eventsXor"
            ],
            "default": "default",
            "title": "Start Type"
        }
    },
    "allOf": [
        {
            "if": {
                "properties": {
                    "type": {
                        "const": "default"
                    }
                }
            },
            "then": SpecialSchemaDefaultStart
        },
        {
            "if": {
                "properties": {
                    "type": {
                        "const": "scheduled"
                    }
                }
            },
            "then": SpecialSchemaScheduledStart
        },
        {
            "if": {
                "properties": {
                    "type": {
                        "const": "event"
                    }
                }
            },
            "then": SpecialSchemaEventStart
        },
        {
            "if": {
                "properties": {
                    "type": {
                        "const": "eventsAnd"
                    }
                }
            },
            "then": SpecialSchemaEventsAndStart
        },
        {
            "if": {
                "properties": {
                    "type": {
                        "const": "eventsXor"
                    }
                }
            },
            "then": SpecialSchemaEventsXorStart
        }
    ]
}

// Functions Schemas
export const FunctionSchemaGlobal = {
    "type": "object",
    "required": [
        "id",
        "service"
    ],
    "properties": {
        "id": {
            "type": "string",
            "title": "ID",
            "description": "Function definition unique identifier."
        },
        "service": {
            "type": "string",
            "title": "Service",
            "description": "The service being referenced."
        }
    }
}

export const FunctionSchemaNamespace = {
    "type": "object",
    "required": [
        "id",
        "service"
    ],
    "properties": {
        "id": {
            "type": "string",
            "title": "ID",
            "description": "Function definition unique identifier."
        },
        "service": {
            "type": "string",
            "title": "Service",
            "description": "The service being referenced."
        }
    }
}

export const FunctionSchemaReusable = {
    "type": "object",
    "required": [
        "id",
        "image"
    ],
    "properties": {
        "id": {
            "type": "string",
            "title": "ID",
            "description": "Function definition unique identifier."
        },
        "image": {
            "type": "string",
            "title": "Image",
            "description": "Image URI.",
            "examples": [
                "direktiv/request",
                "direktiv/python",
                "direktiv/smtp-receiver",
                "direktiv/sql",
                "direktiv/image-watermark"
            ]
        },
        "cmd": {
            "type": "string",
            "title": "CMD",
            "description": "Command to run in container"
        },
        "size": {
            "type": "string",
            "title": "Size",
            "description": "Size of virtual machine"
        },
        "scale": {
            "type": "integer",
            "title": "Scale",
            "description": "Minimum number of instances"
        }
    }
}

export const FunctionSchemaSubflow = {
    "type": "object",
    "required": [
        "id",
        "workflow"
    ],
    "properties": {
        "id": {
            "type": "string",
            "title": "ID",
            "description": "Function definition unique identifier."
        },
        "workflow": {
            "type": "string",
            "title": "Workflow",
            "description": "ID of workflow within the same namespace."
        }
    }
}

//  GenerateFunctionSchemaWithEnum : Generates schema to used for creating new funciton
//  Automatically injects global and namespace service list as enums from arguments
//  Also creates ui-schemas which sets placeholder and whether or not field is readonly (if no services exist)
export function GenerateFunctionSchemaWithEnum(namespaceServices, globalServices, nodes) {
    let nsFuncSchema = FunctionSchemaNamespace
    let globalFuncSchema = FunctionSchemaGlobal
    let subflowFuncSchema = FunctionSchemaSubflow
    let uiSchema = {
        "knative-namespace": {
            "service": {

            }
        },
        "knative-global": {
            "service": {

            }
        }
    }

    if (namespaceServices) {
        if (namespaceServices.length > 0) {
            nsFuncSchema.properties.service.enum = namespaceServices
            uiSchema["knative-namespace"]["service"]["ui:placeholder"] = "Select Service"
        } else {
            delete nsFuncSchema.properties.service.enum
            uiSchema["knative-namespace"]["service"]["ui:placeholder"] = "No Services"
            uiSchema["knative-namespace"]["service"]["ui:readonly"] = true
        }
    }

    if (globalServices) {
        if (globalServices.length > 0) {
            globalFuncSchema.properties.service.enum = globalServices
            uiSchema["knative-global"]["service"]["ui:placeholder"] = "Select Service"
        } else {
            delete globalFuncSchema.properties.service.enum
            uiSchema["knative-global"]["service"]["ui:placeholder"] = "No Services"
            uiSchema["knative-global"]["service"]["ui:readonly"] = true
        }
    }

    subflowFuncSchema.properties.workflow.examples = []
    for (let i = 0; i < nodes.length; i++) {
        if (nodes[i].node.type === "workflow") {
            subflowFuncSchema.properties.workflow.examples.push(nodes[i].node.path)
        }
    }

    return {
        schema: {
            "type": "object",
            "required": [
                "type"
            ],
            "properties": {
                "type": {
                    "enum": [
                        "knative-workflow",
                        "knative-namespace",
                        "knative-global",
                        "subflow"
                    ],
                    "default": "knative-workflow",
                    "title": "Service Type",
                    "description": "Function type of new service"
                }
            },
            "allOf": [
                {
                    "if": {
                        "properties": {
                            "type": {
                                "const": "knative-workflow"
                            }
                        }
                    },
                    "then": FunctionSchemaReusable
                },
                {
                    "if": {
                        "properties": {
                            "type": {
                                "const": "knative-namespace"
                            }
                        }
                    },
                    "then": nsFuncSchema
                },
                {
                    "if": {
                        "properties": {
                            "type": {
                                "const": "knative-global"
                            }
                        }
                    },
                    "then": globalFuncSchema
                },
                {
                    "if": {
                        "properties": {
                            "type": {
                                "const": "subflow"
                            }
                        }
                    },
                    "then": FunctionSchemaSubflow
                }
            ]
        },
        uiSchema: uiSchema
    }
}

export const FunctionSchema = {
    "type": "object",
    "required": [
        "type"
    ],
    "properties": {
        "type": {
            "enum": [
                "knative-workflow",
                "knative-namespace",
                "knative-global",
                "subflow"
            ],
            "default": "knative-workflow",
            "title": "Service Type",
            "description": "Function type of new service"
        }
    },
    "allOf": [
        {
            "if": {
                "properties": {
                    "type": {
                        "const": "knative-workflow"
                    }
                }
            },
            "then": FunctionSchemaReusable
        },
        {
            "if": {
                "properties": {
                    "type": {
                        "const": "knative-namespace"
                    }
                }
            },
            "then": FunctionSchemaNamespace
        },
        {
            "if": {
                "properties": {
                    "type": {
                        "const": "knative-global"
                    }
                }
            },
            "then": FunctionSchemaGlobal
        },
        {
            "if": {
                "properties": {
                    "type": {
                        "const": "subflow"
                    }
                }
            },
            "then": FunctionSchemaSubflow
        }
    ]
}



// Map to all Schemas
export const SchemaMap = {
    // States
    "stateSchemaNoop": StateSchemaNoop,
    "stateSchemaAction": StateSchemaAction,
    "stateSchemaSwitch": StateSchemaSwitch,
    "stateSchemaConsumeEvent": StateSchemaConsumeEvent,
    "stateSchemaDelay": StateSchemaDelay,
    "stateSchemaError": StateSchemaError,
    "stateSchemaEventsAnd": StateSchemaEventsAnd,
    "stateSchemaEventXor": StateSchemaEventXor,
    "stateSchemaForeach": StateSchemaForeach,
    "stateSchemaParallel": StateSchemaParallel,
    "stateSchemaGenerateEvent": StateSchemaGenerateEvent,
    "stateSchemaGetter": StateSchemaGetter,
    "stateSchemaSetter": StateSchemaSetter,
    "stateSchemaValidate": StateSchemaValidate,

    // Functions
    "functionSchemaGlobal": FunctionSchemaGlobal,
    "functionSchemaNamespace": FunctionSchemaNamespace,
    "functionSchemaReusable": FunctionSchemaReusable,
    "functionSchemaSubflow": FunctionSchemaSubflow,
    "functionSchema": FunctionSchema,

    // Special
    "specialSchemaError": SpecialSchemaError,
    "specialStartBlock": SpecialSchemaStart,
}

function functionListToActionEnum(functionList) {
    let availableFunctions = []
    for (let i = 0; i < functionList.length; i++) {
        const f = functionList[i];
        availableFunctions.push(f.id)
    }

    return availableFunctions
}

export function getSchemaDefault(schemaKey) {
    return SchemaMap[schemaKey]
}
export const getSchemaCallbackMap = {
    "stateSchemaAction": (schemaKey, functionList, varList) => {
        let selectedSchema = SchemaMap[schemaKey]
        selectedSchema.properties.action.properties.function.enum = functionListToActionEnum(functionList)
        return selectedSchema
    },
    "stateSchemaForeach": (schemaKey, functionList, varList) => {
        let selectedSchema = SchemaMap[schemaKey]
        selectedSchema.properties.action.properties.function.enum = functionListToActionEnum(functionList)
        return selectedSchema
    },
    "stateSchemaParallel": (schemaKey, functionList, varList) => {
        let selectedSchema = SchemaMap[schemaKey]
        selectedSchema.properties.actions.items.properties.function.enum = functionListToActionEnum(functionList)
        return selectedSchema
    },
    "Default": (schemaKey, functionList, varList) => {
        return getSchemaDefault(schemaKey)
    }
}

// UI Schemas
export const SchemaUIMap = {
    // States
    // "stateSchemaNoop": StateSchemaNoop,
    "stateSchemaAction": {
        "action": {
            "input": {
                "rawYAML": {
                    "ui:widget": "textAreaWidgetYAML"
                },
                "js": {
                    "ui:widget": "textAreaWidgetJS"
                }
            },
            "function": {
                "ui:placeholder": "Select Function"
            }
        },
        "transform": {
            "rawYAML": {
                "ui:widget": "textAreaWidgetYAML"
            },
            "js": {
                "ui:widget": "textAreaWidgetJS"
            }
        }
    },
    // "stateSchemaSwitch": StateSchemaSwitch,
    // "stateSchemaConsumeEvent": StateSchemaConsumeEvent,
    // "stateSchemaDelay": StateSchemaDelay,
    // "stateSchemaError": StateSchemaError,
    // "stateSchemaEventsAnd": StateSchemaEventsAnd,
    "stateSchemaEventXor": {
        "events": {
            "items": {
                "transform": {
                    "rawYAML": {
                        "ui:widget": "textAreaWidgetYAML"
                    },
                    "js": {
                        "ui:widget": "textAreaWidgetJS"
                    }
                }
            }
        },
        "transform": {
            "rawYAML": {
                "ui:widget": "textAreaWidgetYAML"
            },
            "js": {
                "ui:widget": "textAreaWidgetJS"
            }
        }
    },
    "stateSchemaParallel": {
        "actions": {
            "items": {
                "function": {
                    "ui:placeholder": "Select Function"
                }
            }
        }
    },
    "stateSchemaForeach": {
        "action": {
            "input": {
                "rawYAML": {
                    "ui:widget": "textAreaWidgetYAML"
                },
                "js": {
                    "ui:widget": "textAreaWidgetJS"
                }
            },
            "function": {
                "ui:placeholder": "Select Function"
            }
        },
        "transform": {
            "rawYAML": {
                "ui:widget": "textAreaWidgetYAML"
            },
            "js": {
                "ui:widget": "textAreaWidgetJS"
            }
        }
    },
    "stateSchemaGenerateEvent": {
        "event": {
            "data": {
                "rawYAML": {
                    "ui:widget": "textAreaWidgetYAML"
                },
                "js": {
                    "ui:widget": "textAreaWidgetJS"
                }
            }
        },
        "transform": {
            "rawYAML": {
                "ui:widget": "textAreaWidgetYAML"
            },
            "js": {
                "ui:widget": "textAreaWidgetJS"
            }
        }
    },
    // "stateSchemaGetter": StateSchemaGetter,
    "stateSchemaSetter": {
        "variables": {
            "items": {
                "value": {
                    "rawYAML": {
                        "ui:widget": "textAreaWidgetYAML"
                    },
                    "js": {
                        "ui:widget": "textAreaWidgetJS"
                    }
                }
            }
        },
        "transform": {
            "rawYAML": {
                "ui:widget": "textAreaWidgetYAML"
            },
            "js": {
                "ui:widget": "textAreaWidgetJS"
            }
        }
    },
    "stateSchemaValidate": {
        "schema": {
            "ui:widget": "textAreaWidgetYAML"
        },
        "js": {
            "ui:widget": "textAreaWidgetJS"
        }
    },

    // Functions
    // "functionSchemaGlobal": FunctionSchemaGlobal,
    // "functionSchemaNamespace": FunctionSchemaNamespace,
    // "functionSchemaReusable": FunctionSchemaReusable,
    // "functionSchemaSubflow": FunctionSchemaSubflow,
    // "functionSchema": FunctionSchema,

    // Special
    "specialSchemaError": SpecialSchemaError,
}

export const DefaultSchemaUI = {
    "transform": {
        "rawYAML": {
            "ui:widget": "textAreaWidgetYAML"
        },
        "js": {
            "ui:widget": "textAreaWidgetJS"
        }
    }
}
