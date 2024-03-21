import { FC, PropsWithChildren, createContext, useContext } from "react";

import { StateType } from "./type";

const InstanceStateContext = createContext<StateType | null>(null);

const Provider: FC<PropsWithChildren & { instanceId: string }> = ({
  children,
  instanceId,
}) => {
  const value = {
    instanceId,
  };
  return (
    <InstanceStateContext.Provider value={value}>
      {children}
    </InstanceStateContext.Provider>
  );
};

const useInstanceId = () => {
  const context = useContext(InstanceStateContext);
  if (!context) {
    throw new Error("useInstanceId must be used within a InstanceContext");
  }
  return context.instanceId;
};

export { Provider as InstanceStateProvider, useInstanceId };
