import "./style.css";

import FlexBox from "../../components/flexbox";
import { Link } from "react-router-dom";
import React from "react";

function NotFound() {
  return (
    <FlexBox className="center-y center-x col">
      <div>
        <span style={{ fontSize: "120px", fontWeight: "bolder" }}>404</span>
      </div>
      <div style={{ paddingTop: "16px" }}>
        <span>The Page or Resource was not found</span>
      </div>
      <Link to="/" style={{ paddingTop: "8px" }}>
        <div className="link-404">Go back to home</div>
      </Link>
    </FlexBox>
  );
}

export default NotFound;
