import { FC } from "react";
import { NavLink } from "react-router-dom";
import { createClassNames } from "../NavigationLink";
import { pages } from "../../util/router/pages";

const Navigation: FC = () => (
  <>
    {Object.entries(pages).map(([key, item]) => (
      <NavLink
        key={key}
        to={item.createHref()}
        className={({ isActive }) => createClassNames(isActive)}
      >
        <item.icon aria-hidden="true" /> {item.name}
      </NavLink>
    ))}
  </>
);

export default Navigation;
