export type UpdateFilterState = {
  type: "UPDATE_FILTER_STATE_NAME";
  payload: {
    filter: string;
  };
};

export type UpdateFilterWorkflow = {
  type: "UPDATE_FILTER_WORKFLOW";
  payload: {
    filter: string;
  };
};

export type Actions = UpdateFilterState | UpdateFilterWorkflow;
