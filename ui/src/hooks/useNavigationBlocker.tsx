import { BlockerFunction, useBlocker } from "react-router-dom";

import { useEffect } from "react";

/**
 * Define a message to show that text as a warning before leaving the page.
 * Set message to null to disable.
 */
const useNavigationBlocker = (message: string | null) => {
  /**
   * This triggers the browser's built in warning when navigating to another
   * document (that is, leaving this app) with unsaved changes.
   */
  const beforeUnloadHandler = (event: BeforeUnloadEvent) => {
    event.preventDefault();
    event.returnValue = true;
  };

  useEffect(() => {
    if (message) {
      window.addEventListener("beforeunload", beforeUnloadHandler);
    } else {
      window.removeEventListener("beforeunload", beforeUnloadHandler);
    }

    return () =>
      window.removeEventListener("beforeunload", beforeUnloadHandler);
  });

  /**
   * This triggers a confirmation dialog when navigating away from the
   * current route, but staying within our app.
   */
  const blockerFunction: BlockerFunction = ({
    currentLocation,
    nextLocation,
  }) => (message ? currentLocation !== nextLocation : false);

  const blocker = useBlocker(blockerFunction);

  useEffect(() => {
    message && blocker.state === "blocked" && window.confirm(message)
      ? blocker.proceed()
      : blocker.reset?.();
  }, [blocker, blocker.state, message]);
};

export default useNavigationBlocker;
