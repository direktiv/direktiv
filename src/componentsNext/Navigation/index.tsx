import { FC } from "react";
import { NavLink } from "react-router-dom";
import { NavigationLink } from "../NavigationLink";
import { pages } from "../../util/router/pages";

const Navigation: FC = () => (
  <>
    {Object.values(pages).map((item) => (
      <>
        <NavLink key={item.name} to={item.route.path ?? "#"}>
          {item.name} <br />
        </NavLink>
        <NavigationLink key={item.name} href={item.href} active={item.current}>
          <item.icon aria-hidden="true" />
          {item.name}
        </NavigationLink>
      </>
    ))}
  </>
);

export default Navigation;
