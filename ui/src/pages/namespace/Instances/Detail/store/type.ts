import { FiltersObj } from "~/api/logs/query/get";

export type StateType = {
  instanceId: string;
  filters?: FiltersObj;
};

export type UpdateFilterState = {
  type: "UPDATE_FILTER_STATE_NAME";
  payload: {
    stateName: string | undefined;
  };
};

export type UpdateFilterWorkflow = {
  type: "UPDATE_FILTER_WORKFLOW";
  payload: {
    workflowName: string | undefined;
  };
};

export type Actions = UpdateFilterState | UpdateFilterWorkflow;
