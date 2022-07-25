export const NodeStartBlock = {
    name: 'StartBlock',
    family: "special",
    type: "start",
    info: {
        requiresInit: false,
        actions: true,
        description: "DEBUG",
        longDescription: `DEBUG`,
        link: ""
    },
    data: {
        schemaKey: 'specialStartBlock',
        formData: {}
    },
    connections: {
        input: 0,
        output: 1
    },
    html: 'Start Block'
}

export const NodeErrorBlock = {
    name: 'CatchError',
    family: "special",
    type: "catch",
    info: {
        requiresInit: false,
        actions: true,
        description: "Catches errors and routes the flow of the workflow depending on the error type.",
        longDescription: `DEBUG`,
        link: null
    },
    data: {
        schemaKey: 'specialSchemaError',
        formData: {}
    },
    connections: {
        input: 1,
        output: 0
    },
    html: 'Catch Error'
}


export const NodeStateAction = {
    name: 'StateAction',
    family: "primitive",
    type: "action",
    info: {
        requiresInit: true,
        actions: true,
        description: "The Action State runs another workflow as a subflow, or a function as defined in the forms action definition",
        longDescription: ``,
        link: "https://docs.direktiv.io/v0.6.0/specification/#actionstate"
    },
    data: {
        schemaKey: 'stateSchemaAction',
        formData: {}
    },
    connections: {
        input: 1,
        output: 2
    },
    html: 'Action State'
}



export const ActionsNodes = [
    {
        name: 'StateNoop',
        family: "primitive",
        type: "noop",
        info: {
            requiresInit: true,
            actions: true,
            description: "The No-op State exists for when nothing more than generic state functionality is required.",
            longDescription: `The No-op State exists for when nothing more than generic state functionality is required. A common use-case would be to perform a jq operation on the state data without performing another operation.`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#noopstate"
        },
        data: {
            schemaKey: 'stateSchemaNoop',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Noop State'
    },
    {
        name: 'StateConsumeEvent',
        family: "primitive",
        type: "consumeEvent",
        info: {
            requiresInit: true,
            actions: true,
            description: "The ConsumeEvent State is the simplest state you can use to listen for CloudEvents in the middle of a workflow.",
            longDescription: `todo`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#consumeeventstate"
        },
        data: {
            schemaKey: 'stateSchemaConsumeEvent',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Consume Event State'
    },
    {
        name: 'StateDelay',
        family: "primitive",
        type: "delay",
        info: {
            requiresInit: true,
            actions: true,
            description: "The Delay State pauses execution of the workflow for a predefined length of time.",
            longDescription: `todo`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#delaystate"
        },
        data: {
            schemaKey: 'stateSchemaDelay',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Delay State'
    },
    {
        name: 'StateError',
        family: "primitive",
        type: "error",
        info: {
            requiresInit: true,
            actions: true,
            description: "The Error State allows a subflow to throw an error, catchable by the calling workflow.",
            longDescription: `todo`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#errorstate"
        },
        data: {
            schemaKey: 'stateSchemaError',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Error State'
    },
    {
        name: 'StateEventsAnd',
        family: "primitive",
        type: "eventsAnd",
        info: {
            requiresInit: true,
            actions: true,
            description: "Halts execution until it receives a matching event for every event in defined in its events.",
            longDescription: `todo`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#eventandstate"
        },
        data: {
            schemaKey: 'stateSchemaEventsAnd',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'EventsAnd State'
    },
    {
        name: 'StateEventXor',
        family: "primitive",
        type: "eventsXor",
        info: {
            requiresInit: true,
            actions: true,
            description: "Halts execution until it receives any matching event defined in its events.",
            longDescription: `todo`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#eventxorstate"
        },
        data: {
            schemaKey: 'stateSchemaEventXor',
            formData: {}
        },
        connections: {
            input: 1,
            output: 1 //TODO: Special case Error output is in node 1
        },
        html: 'EventXor State'
    },
    {
        name: 'StateForeach',
        family: "primitive",
        type: "foreach",
        info: {
            requiresInit: true,
            actions: true,
            description: "The ForeachState can be used to split up state data into an array and then perform an action on each element in parallel.",
            longDescription: `todo`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#foreachstate"
        },
        data: {
            schemaKey: 'stateSchemaForeach',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Foreach State'
    },
    {
        name: 'StateGenerateEvent',
        family: "primitive",
        type: "generateEvent",
        info: {
            requiresInit: true,
            actions: true,
            description: "The GenerateEvent State will produce an event that other workflows could listen for.",
            longDescription: `todo`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#generateeventstate"
        },
        data: {
            schemaKey: 'stateSchemaGenerateEvent',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Generate Event State'
    },
    {
        name: 'StateGetter',
        family: "primitive",
        type: "getter",
        info: {
            requiresInit: true,
            actions: true,
            description: "The getter state is used to retrieve persistent data.",
            longDescription: `todo`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#getterstate"
        },
        data: {
            schemaKey: 'stateSchemaGetter',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Getter State'
    },
    {
        name: 'StateSetter',
        family: "primitive",
        type: "setter",
        info: {
            requiresInit: true,
            actions: true,
            description: "The setter state is used to store persistent data. ",
            longDescription: `todo`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#setterstate"
        },
        data: {
            schemaKey: 'stateSchemaSetter',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Setter State'
    },
    NodeStateAction,
    {
        name: 'StateValidate',
        family: "primitive",
        type: "validate",
        info: {
            requiresInit: true,
            actions: true,
            description: "The Validate State can be used to validate the structure of the state's data using a JSON schema.",
            longDescription: `todo`,
            link: "https://docs.direktiv.io/v0.6.0/specification/#validatestate"
        },
        data: {
            schemaKey: 'stateSchemaValidate',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Validate State'
    },
    {
        name: 'StateSwitch',
        family: "primitive",
        type: "switch",
        info: {
            requiresInit: true,
            actions: true,
            description: "The Switch State is used to perform conditional transitions based on the current state information",
            longDescription: ``,
            link: "https://docs.direktiv.io/v0.6.0/specification/#switchstate"
        },
        data: {
            schemaKey: 'stateSchemaSwitch',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Switch State'
    },
    {
        name: 'StateParallel',
        family: "primitive",
        type: "parallel",
        info: {
            requiresInit: true,
            actions: true,
            description: "The Switch State is used to perform conditional transitions based on the current state information",
            longDescription: ``,
            link: "https://docs.direktiv.io/v0.6.0/specification/#parallelstate"
        },
        data: {
            schemaKey: 'stateSchemaParallel',
            formData: {}
        },
        connections: {
            input: 1,
            output: 2
        },
        html: 'Parallel State'
    },
    NodeErrorBlock
]
