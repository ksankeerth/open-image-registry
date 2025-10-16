import React from "react";
import logo from "./../assets/logo.png";
import { classNames } from "primereact/utils";

type LogoProps = {
  showNameInOneLine: boolean;
  // showNameAndLogoInOneLine: boolean;
};

const LogoComponent = (props: LogoProps) => {
  return (
    <div className="flex-grow-1 flex align-items-center	">
      <div>
        <img src={logo} className="h-4rem" />
      </div>

      <div
        className={classNames(
          "flex",
          props.showNameInOneLine ? "flex-row" : "flex-column"
        )}
      >
        <div
          className="text-green-300"
          style={{
            fontFamily: "Major Mono Display, monospace",
            fontSize: props.showNameInOneLine ? 24 : 18,
          }}
        >
          OPEN IMAGE
        </div>
        {props.showNameInOneLine && <span>&nbsp;&nbsp;&nbsp;</span>}
        <div
          style={{
            fontFamily: "Major Mono Display, monospace",
            fontSize: props.showNameInOneLine ? 24 : 20,
            color: "#007700",
          }}
        >
          REGISTRY
        </div>
      </div>
    </div>
  );
};

export default LogoComponent;