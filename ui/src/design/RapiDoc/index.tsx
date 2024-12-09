import "./styles/RapiDoc.css";
import "rapidoc";

import { twMergeClsx } from "~/util/helpers";

interface RapiDocProps {
  spec: object;
  className?: string;
}

export function RapiDoc({ spec, className }: RapiDocProps) {
  return (
    <div className={twMergeClsx("size-full overflow-scroll", className)}>
      <rapi-doc
        ref={(rapiDocElement) => rapiDocElement?.loadSpec(spec)}
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
