import React from 'react';
import logo from './../assets/logo.png';
import { classNames } from 'primereact/utils';

type LogoProps = {
  showNameInOneLine: boolean;
};

const LogoComponent = (props: LogoProps) => {
  const baseTextStyle: React.CSSProperties = {
    fontFamily: 'Montserrat, sans-serif',
    lineHeight: 1.1,
  };

  return (
    <div className="flex align-items-center gap-2">
      <img src={logo} className="h-4rem" alt="Logo" />

      <div
        className={classNames(
          'flex',
          props.showNameInOneLine ? 'flex-row align-items-center' : 'flex-column'
        )}
      >
        <div
          style={{
            ...baseTextStyle,
            fontWeight: 500,
            fontSize: props.showNameInOneLine ? 22 : 16,
            color: '#9ACD7A',
          }}
        >
          OPEN IMAGE
        </div>

        {props.showNameInOneLine && <span className="mx-1" />}

        <div
          style={{
            ...baseTextStyle,
            fontWeight: 700,
            fontSize: props.showNameInOneLine ? 22 : 18,
            color: '#007700',
            letterSpacing: '0.5px',
          }}
        >
          REGISTRY
        </div>
      </div>
    </div>
  );
};

export default LogoComponent;
