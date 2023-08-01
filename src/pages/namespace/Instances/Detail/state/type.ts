export type UpdateFilterState = {
  type: "UPDATE_FILTER_STATE_NAME";
  payload: {
    stateName: string;
  };
};

export type UpdateFilterWorkflow = {
  type: "UPDATE_FILTER_WORKFLOW";
  payload: {
    workflowName: string;
  };
};

export type Actions = UpdateFilterState | UpdateFilterWorkflow;
