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

/**
 * Gets the secrets specified in the argument.
 * @param keys an array of secret ids.
 */
// Not currently implemented
// declare function getSecrets<T extends readonly string[]>(
//   keys: readonly [...T]
// ): { [K in T[number]]: string };

/**
 * Returns the instance id of the workflow
 */
declare function id(): string;

declare type DateObject = {
  /**
   * Get unix time from a date object.
   * @returns seconds since 1970-01-01.
   */
  unix: () => number;
  /**
   * Get formatted string from a date object. Use golang time
   * format string, e.g., "2006-01-02 15:30".
   * @returns formatted date/time.
   */
  format: (template: string) => string;
};

/**
 * Returns a date object which can be used to format time.
 * This object can be used with now().unix()to get the seconds
 * since 1.1.1970 and now().format("2006-01-02 15:30")
 */
declare function now(): DateObject;

declare type ActionConfig = {
  type: string; // todo: enum
  image: string;
  envs: {
    name: string;
    value: string;
  }[];
};

/**
 * Creates a custom action that can then be called as a
 * typescript function.
 * @param ActionConfig
 */
declare function generateAction(ActionConfig): () => void;
