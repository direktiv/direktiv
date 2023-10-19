import "tippy.js/dist/tippy.css";
import "./style.css";

import React from "react";
import Tippy from "@tippyjs/react";
import { VscInfo } from "react-icons/vsc";

export interface HelpIconProps {
  /**
   * Tooltip message.
   */
  msg: string;
  /**
   * z-index of component.
   */
  zIndex: number;
}

/**
 * Simple Help Icon that will display a message as a tooltip on hover.
 */
function HelpIcon({
  msg = "No help text provided",
  zIndex = 10,
}: HelpIconProps) {
  return (
    <>
      <Tippy content={msg} trigger="mouseenter focus click" zIndex={zIndex}>
        <div className="iconWrapper">
          <VscInfo className="grey-text" />
        </div>
      </Tippy>
    </>
  );
}

export default HelpIcon;
