import { ShouldBlockFn, useBlocker } from "@tanstack/react-router";

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

  const shouldBlockFn: ShouldBlockFn = ({ current, next }) =>
    message ? current !== next : false;

  const { status, proceed, reset } = useBlocker({
    shouldBlockFn,
    withResolver: true,
  });

  useEffect(() => {
    if (!message) {
      return proceed?.();
    }
    if (status === "blocked" && window.confirm(message)) {
      return proceed();
    }
  }, [message, status, proceed, reset]);
};

export default useNavigationBlocker;
