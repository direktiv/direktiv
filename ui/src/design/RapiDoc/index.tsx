import "./styles/RapiDoc.css";
import "rapidoc";

import { twMergeClsx } from "~/util/helpers";

export function RapiDoc({
  spec,
  className,
}: {
  spec: object;
  className?: string;
}) {
  return (
    <div
      className={twMergeClsx(
        "my-1 lg:h-[calc(100vh-240px)] sm:h-[calc(100vh-320px)]",
        className
      )}
    >
      <rapi-doc
        ref={(rapiDocElement) => rapiDocElement?.loadSpec(spec)}
        id="rapidoc"
        render-style="view" // (view | read) sidebar visibility
        allow-try="true" // Test Endpoints
        allow-authentication="false"
        show-header="false"
        show-info="true"
        theme="light"
        schema-style="table" // table | tree
        primary-color="#5364FF"
      ></rapi-doc>
    </div>
  );
}
