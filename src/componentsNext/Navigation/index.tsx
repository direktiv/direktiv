import { FC } from "react";
import { NavLink } from "react-router-dom";
import { createClassNames } from "../NavigationLink";
import { pages } from "../../util/router/pages";

const Navigation: FC = () => (
  <>
    {Object.values(pages).map((item) => (
      <NavLink
        key={item.name}
        to={item.route.path ?? "#"}
        className={({ isActive }) => createClassNames(isActive)}
      >
        <item.icon aria-hidden="true" /> {item.name}
      </NavLink>
    ))}
  </>
);

export default Navigation;
