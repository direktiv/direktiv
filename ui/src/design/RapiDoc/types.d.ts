declare module "~/design/RapiDoc" {
  interface RapiDocProps {
    spec?: any;
    className?: string;
  }

  export function RapiDoc(props: RapiDocProps): JSX.Element;
}

interface RapiDocElement extends HTMLElement {
  loadSpec(spec: object): void;
}

declare namespace JSX {
  interface IntrinsicElements {
    "rapi-doc": {
      id?: string;
      "spec-url"?: string;
      spec?: string | object;
      "render-style"?: string;
      "show-header"?: string;
      "show-info"?: string;
      theme?: string;
      "primary-color"?: string;
      "font-family"?: string;
    };
  }
}

declare module "*.yaml" {
  const content: any;
  export default content;
}
