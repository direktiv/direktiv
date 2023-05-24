import { FC, useEffect, useState } from "react";

import { useVars } from "~/api/vars/query/get";

const VarsList: FC = () => {
  const data = useVars();

  const vars = data.data?.variables?.results ?? null;

  return (
    <ul>
      {vars?.map((item, i) => (
        <li key={i}>{item.name}</li>
      ))}
    </ul>
  );
};

export default VarsList;
