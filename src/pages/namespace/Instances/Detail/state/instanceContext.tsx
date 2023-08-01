import {
  FC,
  PropsWithChildren,
  createContext,
  useContext,
  useReducer,
} from "react";

import { FiltersObj } from "~/api/logs/query/get";
import { Actions as InstanceActions } from "./type";

type StateType = {
  filters: FiltersObj;
};

const defaultState = {
  filters: {},
};

const InstanceStateContext = createContext<StateType>(defaultState);

const InstanceDispatchContext =
  createContext<React.Dispatch<InstanceActions> | null>(null);

function stateReducer(state: StateType, action: InstanceActions) {
  switch (action.type) {
    case "UPDATE_FILTER_STATE_NAME": {
      return {
        ...state,
        filters: {
          ...state.filters,
        },
      };
    }
    case "UPDATE_FILTER_WORKFLOW": {
      return {
        ...state,
        filters: {
          ...state.filters,
        },
      };
    }

    default: {
      return state;
    }
  }
}

const Provider: FC<PropsWithChildren> = ({ children }) => {
  const [state, dispatch] = useReducer(stateReducer, defaultState);

  return (
    <InstanceStateContext.Provider value={state}>
      <InstanceDispatchContext.Provider value={dispatch}>
        {children}
      </InstanceDispatchContext.Provider>
    </InstanceStateContext.Provider>
  );
};

const useFilters = () => {
  const context = useContext(InstanceStateContext);
  if (context === undefined) {
    throw new Error("useFilter must be used within a InstanceContext");
  }
  return context.filters;
};

const useActions = () => {
  const context = useContext(InstanceDispatchContext);
  if (!context) {
    throw new Error("useActions must be used within a InstanceDispatchContext");
  }

  return {
    updateFilterStateName: (stateName: string) => {
      context({
        type: "UPDATE_FILTER_STATE_NAME",
        payload: {
          stateName,
        },
      });
    },
    updateFilterWorkflow: (workflowName: string) => {
      context({
        type: "UPDATE_FILTER_WORKFLOW",
        payload: {
          workflowName,
        },
      });
    },
  };
};

export { Provider as InstanceStateProvider, useFilters, useActions };
