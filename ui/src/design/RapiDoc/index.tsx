import "./styles/RapiDoc.css";

import exampleSpec from "~/design/RapiDoc/example.json";
import { useEffect } from "react";

export function RapiDoc() {
  useEffect(() => {
    const docEl = document.getElementById("rapidoc") as RapiDocElement;
    if (docEl) {
      docEl.loadSpec(exampleSpec);
    }
  }, []);

  Hello;

  return (
    <div style={{ height: "100%", width: "100%", overflow: "auto" }}>
      <rapi-doc
        id="rapidoc"
        render-style="read"
        allow-try="false"
        allow-authentication="false"
        show-header="false"
        show-info="true"
        theme="light"
        primary-color="#5364FF"
      ></rapi-doc>
    </div>
  );
}
