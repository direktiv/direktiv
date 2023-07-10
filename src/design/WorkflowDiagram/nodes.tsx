import { Handle, Position } from "reactflow";

interface StateProps {
  data: {
    label: string;
    type: string;
  };
}

export function State(props: StateProps) {
  const { data } = props;
  const { label, type } = data;
  return (
    <div className="state">
      <Handle type="target" position={Position.Left} id="default" />
      <div>
        <div>{type}</div>
      </div>
      <h2>{label}</h2>
      <Handle type="source" position={Position.Right} id="default" />
    </div>
  );
}

export function Start() {
  return (
    <div className="normal">
      <Handle type="source" position={Position.Right} />
      <div className="start" />
    </div>
  );
}

export function End() {
  return (
    <div className="normal">
      <div className="end" />
      <Handle type="target" position={Position.Left} />
    </div>
  );
}
