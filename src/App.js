import './App.css';
import './util/responsive.css';
import MainLayout from './layouts/main';
import FlexBox from './components/flexbox';
import { VscSignOut } from 'react-icons/vsc';
import {NavItem} from './components/navbar'
function App() {
    return (
      <div className="App">
        <MainLayout footer={
            <>
                    <FlexBox>
                        <FlexBox className="nav-items" style={{ paddingLeft: "10px" }}>
                            <ul style={{ marginTop: "0px" }}>
                                <li>
                                    <NavItem className="red-text" label="Log Out">
                                        <VscSignOut />
                                    </NavItem>
                                </li>
                            </ul>
                        </FlexBox>
                    </FlexBox>

                    <div>
                        <FlexBox className="col navbar-userinfo">
                            {/* <FlexBox className="navbar-username">
                                UserName007
                            </FlexBox> */}
                            <FlexBox className="navbar-version">
                                Version: 0.5.8 (abdgdj)
                            </FlexBox>
                        </FlexBox>
                    </div>
            </>
        } extraRoutes={[]} extraNavigation={[]}/>
      </div>
  );
}

export default App;
