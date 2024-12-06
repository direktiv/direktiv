import "./styles/RapiDoc.css";
import "rapidoc";

import { useEffect } from "react";

interface RapiDocProps {
  spec: string | object;
  className?: string;
}

export function RapiDoc({ spec, className }: RapiDocProps) {
  useEffect(() => {
    const docEl = document.getElementById("rapidoc") as RapiDocElement;
    if (docEl) {
      if (typeof spec === "string") {
        docEl.setAttribute("spec-url", spec);
      } else {
        docEl.loadSpec(spec);
      }
    }
  }, [spec]);

  return (
    <div className={`${className} size-full overflow-scroll`}>
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
