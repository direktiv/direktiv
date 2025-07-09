import { HTMLAttributes, createContext, useContext, useState } from "react";

import { twMergeClsx } from "~/util/helpers";

type LocalDialogContainerContextValue = {
  container: HTMLDivElement | null;
};

const LocalDialogContainerContext =
  createContext<LocalDialogContainerContextValue | null>(null);

LocalDialogContainerContext.displayName = "LocalDialogContainerContext";

export const useLocalDialogContainer = () => {
  const context = useContext(LocalDialogContainerContext);
  if (!context)
    throw new Error(
      "useLocalDialogContainer must be used inside <LocalDialogContainer>"
    );
  return context;
};

type LocalDialogContainerProps = HTMLAttributes<HTMLDivElement>;

export const LocalDialogContainer = ({
  children,
  className,
}: LocalDialogContainerProps) => {
  const [container, setContainer] = useState<HTMLDivElement | null>(null);

  return (
    <LocalDialogContainerContext.Provider value={{ container }}>
      <div ref={setContainer} className={twMergeClsx("relative", className)}>
        {children}
      </div>
    </LocalDialogContainerContext.Provider>
  );
};
