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

/**
 * Config for action
 */
declare type ActionConfig = {
  image: string;
  type?: "workflow" | "namespace" | "system";
  size?: "small" | "medium" | "large";
  retries?: number;
  body?: object;
  envs?: {
    name: string;
    value: string;
  }[];
};

/**
 * Creates a custom action that can then be called as a
 * typescript function.
 *
 * @param ActionConfig configuration object
 * - image: required, image to run as a container
 * - type: optional, defaults to "local", or "namespace" | "system"
 * - size: optional, defaults to "small", or "medium" | "large"
 * - retries: optional, number,
 * - cmd: optional, cmd to run in the container
 * - envs: optional, { name: string, value: string }[].
 */
declare function generateAction(
  config: ActionConfig
): (payload: unknown) => void;

/**
 * Returns a map where the key is the secret name and the value is the value of the secret
 * @param secrets the array of secrets names
 */
declare function getSecrets(secrets: string[]): Record<string, string>;

/**
 * Returns the secret value
 * @param secret the name of one secret
 */
declare function getSecret(secret: string): string;
