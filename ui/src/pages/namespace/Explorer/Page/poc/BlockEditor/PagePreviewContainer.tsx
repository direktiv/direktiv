import { PropsWithChildren, useEffect, useRef } from "react";

import { LocalDialogContainer } from "~/design/LocalDialog/container";
import { usePageStateContext } from "../PageCompiler/context/pageCompilerContext";

export const PagePreviewContainer = ({ children }: PropsWithChildren) => {
  const scrollRef = useRef<HTMLDivElement>(null);
  const { setScrollPos } = usePageStateContext();

  useEffect(() => {
    const element = scrollRef.current;
    if (!element) return;

    const onScroll = () => setScrollPos(element.scrollTop);

    element.addEventListener("scroll", onScroll);
    return () => element.removeEventListener("scroll", onScroll);
  }, [setScrollPos]);

  return (
    <div className="lg:overflow-y-auto" ref={scrollRef}>
      <LocalDialogContainer className="min-w-0 flex-1">
        <div className="mx-auto min-h-[55vh] max-w-screen-lg overflow-hidden p-4">
          {children}
        </div>
      </LocalDialogContainer>
    </div>
  );
};
