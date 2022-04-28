import * as React from "react";
import {
  UNSAFE_NavigationContext as NavigationContext
} from "react-router-dom";

export function useBlocker(blocker, when = true) {
  const { navigator } = React.useContext(NavigationContext);
  React.useEffect(() => {
    if (!when) return;const unblock = navigator.block((tx) => {
      const autoUnblockingTx = {
        ...tx,
        retry() {
          unblock();
          tx.retry();
        }
      };blocker(autoUnblockingTx);
    });return unblock;
  }, [navigator, blocker, when]);
}

export default function usePrompt(message, when = true) {
  const blocker = React.useCallback(
    (tx) => {
      if (windowBlocker(message)) tx.retry();
    },
    [message]
  );useBlocker(blocker, when);
}

export function windowBlocker(msg) {
  const windowMsg = msg ? msg : "Are you sure you want to exit page?"
  return window.confirm(windowMsg)
}