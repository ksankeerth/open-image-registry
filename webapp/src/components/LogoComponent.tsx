import React from "react";
import logo from "./../assets/logo.png";

const LogoComponent = () => {
  return (
    <div className="flex-grow-1 flex align-items-center	">
      <img src={logo} className="h-4rem" />
      <div className="flex flex-column">
        <div
          className="text-green-300"
          style={{
            fontFamily: "Major Mono Display, monospace",
            fontSize: 18,
          }}
        >
          OPEN IMAGE
        </div>
        <div
          style={{
            fontFamily: "Major Mono Display, monospace",
            fontSize: 20,
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
