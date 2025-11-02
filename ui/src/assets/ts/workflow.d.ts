type FlowDefinition = {
  type: "default";
  timeout: string;
  state: string;
};

type StateFunction<T> = (params: T) => void;

/**
 * Will transition to the next workflow state.
 * @param stateFn state function to run next, e.g., stateSecond.
 * @param stateFnParams params passed into the next state function.
 */
declare function transition<T>(stateFn: StateFunction<T>, stateFnParams: T);

/**
 * Will complete the workflow, returning the result.
 * @param data end result output by the workflow, usually a JSON object.
 */
declare function finish<T>(data: T);

/**
 * Waits for a number of seconds.
 * @param seconds time to wait.
 */
declare function sleep(seconds: number): void;
