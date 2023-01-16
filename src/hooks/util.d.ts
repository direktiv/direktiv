interface ListStateAction {
    type: string,
    data: object,
}  

export const StateReducer: (
    state: any,
    action: ListStateAction
) => any;


interface EventStatAction {
    event: string,
    data: object,

    idKey ?: string
    idNewItemKey? :string,
    idData?: string,
}  

export const EventStateReducer: (
    state: any,
    action: EventStatAction
) => any;