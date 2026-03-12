import React, { FC, PropsWithChildren } from "react";

import { Elbows } from "../Elbows";

const Connector = () => (
  <svg width="32" height="64" className="shrink-0">
    <line x1="0" y1="32" x2="32" y2="32" className="stroke-gray-400 stroke-2" />
  </svg>
);

const Filler = () => (
  <svg
    viewBox="0 0 64 64"
    preserveAspectRatio="none"
    className="my-[16px] h-[64px] w-0 shrink-0 grow"
  >
    <line
      x1="0"
      y1="32"
      x2="100"
      y2="32"
      className="stroke-gray-400 stroke-2"
    />
  </svg>
);

const InvertedFiller = () => (
  <svg
    viewBox="0 0 64 64"
    preserveAspectRatio="none"
    className="my-[16px] h-[64px] w-5 shrink-0"
  >
    <line
      x1="0"
      y1="32"
      x2="100"
      y2="32"
      className="stroke-gray-400 stroke-2"
    />
  </svg>
);

type OrGroupProps = PropsWithChildren & {
  childSizes: number[];
};

const AndGroup: FC<PropsWithChildren> = ({ children }) => {
  const childrenArray = React.Children.toArray(children);

  return (
    <div className="flex flex-row items-center">
      <InvertedFiller />
      {childrenArray.map((item, index) => (
        <React.Fragment key={index}>{item}</React.Fragment>
      ))}
      <Filler />
    </div>
  );
};

const OrGroup: FC<OrGroupProps> = ({ children, childSizes }) => {
  const childrenArray = React.Children.toArray(children);

  return (
    <div className="flex flex-row">
      <Elbows targets={childSizes} />
      <div className="flex flex-col">{childrenArray}</div>
      <Elbows targets={childSizes} reverse={true} />
    </div>
  );
};

export { AndGroup, Connector, OrGroup };
