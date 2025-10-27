import { useEffect, useState } from "react";

import { useLocalDialogContainer } from "./container";

export const useDialogContainerWrapper = () => {
  const { container } = useLocalDialogContainer();
  const [dialogContainerWrapper, setDialogContainerWrapper] = useState<
    HTMLElement | null | undefined
  >(undefined);

  useEffect(() => {
    const updateRect = () => {
      setDialogContainerWrapper(container?.parentElement);
    };
    updateRect();
    window.addEventListener("resize", updateRect);
    return () => window.removeEventListener("resize", updateRect);
  }, [container]);

  return dialogContainerWrapper;
};
