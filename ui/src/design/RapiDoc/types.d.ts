// These declarations are needed because we're using a custom HTML element <rapi-doc

interface RapiDocElement extends HTMLElement {
  loadSpec(spec: object): void;
}
declare namespace JSX {
  interface IntrinsicElements {
    "rapi-doc": {
      id?: string;
      theme?: string;
    };
  }
}

// This declaration is needed when importing YAML files
// js-yaml /  @types/js-yaml  / vite-plugin-yaml
declare module "*.yaml" {
  const content: Record<string, unknown>;
  export default content;
}
