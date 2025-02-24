import "./styles/RapiDoc.css";
import "rapidoc";

import { twMergeClsx } from "~/util/helpers";

interface BaseSpec {
  info: { title: string; version: string };
}

interface OpenApiSpec extends BaseSpec {
  openapi: string;
}

interface SwaggerSpec extends BaseSpec {
  swagger: string;
}

type Spec = OpenApiSpec | SwaggerSpec;

export function RapiDoc({
  spec,
  className,
}: {
  spec: Spec;
  className?: string;
}) {
  return (
    <div className={twMergeClsx("size-full", className)}>
      <rapi-doc
        ref={(rapiDocElement) => rapiDocElement?.loadSpec(spec)}
        id="rapidoc"
        render-style="view" // (view | read) sidebar visibility
        allow-try="true" // Add a "Try" button to execute requests
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
