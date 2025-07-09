import { HTMLAttributes, createContext, useContext, useState } from "react";

import { twMerge } from "tailwind-merge";

const LocalDialogContainerContext = createContext<HTMLDivElement | null>(null);

export const useLocalDialogContainer = () => {
  const context = useContext(LocalDialogContainerContext);
  if (!context) throw new Error("Must be used inside <LocalDialogContainer>");
  return context;
};

type LocalDialogContainerProps = HTMLAttributes<HTMLDivElement>;

export const LocalDialogContainer = ({
  children,
  className,
}: LocalDialogContainerProps) => {
  const [container, setContainer] = useState<HTMLDivElement | null>(null);

  return (
    <LocalDialogContainerContext.Provider value={container}>
      <div ref={setContainer} className={twMerge("relative", className)}>
        {children}
      </div>
    </LocalDialogContainerContext.Provider>
  );
};
