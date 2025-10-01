type FlowDefinition = {
  type: "default";
  timeout: string;
  state: string;
};

/**
 * Will transition to the next workflow state.
 * @param params
 */
declare function transition(fn: WorkflowState, data: unknown);

/**
 * Will transition to the next workflow state.
 * @param params
 */
declare function finish(data: unknown);

/**
 * Example method (outdated, just here for demo)
 * @param params
 */
declare function getFile(params: {
  /**
   * File name
   */
  name: string;
  /**
   * Permission
   */
  permission: number;
  /**
   * What is this for exaclty? What are the possible values?
   */
  scope: "shared" | "other";
}): void;
