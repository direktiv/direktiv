import React from "react";
import { useEffect } from "react";

export const DEFAULT_THEME = "light";

export default (Story, context) => {
  const { theme } = context.globals;

  useEffect(() => {
    const htmlTag = document.documentElement;
    // Set the "data-theme" attribute on the iFrame html tag
    htmlTag.setAttribute("data-theme", theme || DEFAULT_THEME);
  }, [theme]);

  return <Story />;
};
