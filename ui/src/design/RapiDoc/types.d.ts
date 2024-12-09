// These declarations are needed because we're using a custom HTML element <rapi-doc

interface RapiDocElement extends HTMLElement {
  loadSpec(spec: string | object): void;
}
declare namespace JSX {
  interface IntrinsicElements {
    "rapi-doc": {
      id?: string;
      theme?: string;
      ref?: React.Ref<RapiDocElement>;
    };
  }
}
