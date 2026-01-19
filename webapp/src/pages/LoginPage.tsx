import React, { useEffect, useState } from 'react';
import './login.css';
import LogoComponent from '../components/LogoComponent';
import { InputText } from 'primereact/inputtext';
import { Button } from 'primereact/button';
import { Checkbox } from 'primereact/checkbox';
import { ProgressSpinner } from 'primereact/progressspinner';
import { useNavigate } from 'react-router-dom';
import { postAuthLogin } from '../api';

const LoginPage = () => {
  const [username, setUsername] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  const [rememberMe, setRememberMe] = useState<boolean>(false);
  const [showBackdrop, setShowBackdrop] = useState<boolean>(false);

  const [processing, setProcessing] = useState<boolean>(false);

  const [errorMsg, setErrorMsg] = useState<string>('');

  const navigate = useNavigate();

  useEffect(() => {
    if (!processing) {
      const timer = setTimeout(() => {
        setShowBackdrop(false);
      }, 200);
      return () => clearTimeout(timer);
    } else {
      setShowBackdrop(true);
    }
  }, [processing]);

  const handleLogin = async () => {
    setProcessing(true);

    const { data, error } = await postAuthLogin({
      body: {
        username,
        password,
      },
    });

    if (error) {
      setProcessing(false);
      setErrorMsg(error.error_message);
    }

    if (data) {
      navigate('/');
    }
  };

  return (
    <div className="flex flex-row min-h-screen max-h-screen">
      <div className="w-6 login-left-container">
        <div className="animation-container">
          {/* Sky with clouds */}
          <div className="sky">
            <div className="cloud cloud1"></div>
            <div className="cloud cloud2"></div>
          </div>

          {/* Floating Particles for atmosphere */}
          <div className="particles">
            <div className="particle"></div>
            <div className="particle"></div>
            <div className="particle"></div>
            <div className="particle"></div>
            <div className="particle"></div>
          </div>

          {/* Seagulls */}
          <div className="seagull seagull1"></div>
          <div className="seagull seagull2"></div>

          {/* Water */}
          <div className="water">
            <div className="wave"></div>
          </div>

          {/* Dock */}
          <div className="dock"></div>

          {/* Ship with containers */}
          <div className="ship">
            <div className="ship-smoke">
              <div className="smoke-puff"></div>
              <div className="smoke-puff"></div>
              <div className="smoke-puff"></div>
            </div>
            <div className="ship-chimney"></div>
            <div className="ship-containers">
              <div className="container-stack">
                <div className="mini-container"></div>
                <div className="mini-container"></div>
              </div>
              <div className="container-stack">
                <div className="mini-container"></div>
                <div className="mini-container"></div>
              </div>
            </div>
            <div className="ship-cabin"></div>
            <div className="ship-body"></div>
          </div>

          {/* Crane */}
          <div className="crane">
            <div className="crane-base"></div>
            <div className="crane-tower">
              <div className="crane-arm">
                <div className="crane-hook">
                  <div className="crane-container"></div>
                </div>
              </div>
            </div>
          </div>

          {/* Harbor storage containers */}
          <div className="harbor-storage">
            <div className="stored-container"></div>
            <div className="stored-container"></div>
            <div className="stored-container"></div>
            <div className="stored-container"></div>
            <div className="stored-container"></div>
            <div className="stored-container"></div>
            <div className="stored-container"></div>
            <div className="stored-container"></div>
            <div className="stored-container"></div>
            <div className="stored-container"></div>
            <div className="stored-container"></div>
            <div className="stored-container"></div>
          </div>
        </div>
      </div>
      <div className="w-6 flex align-items-center justify-content-center relative">
        {showBackdrop && (
          <div
            className="fixed top-0 left-50 bottom-0 h-full w-6 surface-50 opacity-70 flex align-items-center justify-content-center"
            style={{ zIndex: 1000 }}
          >
            <div className="flex flex-column align-items-center">
              <ProgressSpinner style={{ width: '50px', height: '50px' }} />
            </div>
          </div>
        )}

        <div className="flex flex-column">
          <div className="flex flex-row justify-content-center gap-2">
            <span
              style={{
                // fontFamily: "Major Mono Display, monospace",
                fontSize: 20,
                color: '#007700',
              }}
            >
              Welcome back
            </span>
          </div>
          <LogoComponent showNameInOneLine={true} />
          <div className="flex flex-row justify-content-center text-color font-medium text-sm">
            Sign in to continue
          </div>
          <div className="p-4"></div>
          <div className="flex flex-column gap-4">
            <div className="flex flex-column gap-2">
              <label htmlFor="username" className="text-color font-medium text-md">
                Username
              </label>
              <InputText
                id="username"
                size={45}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>
            <div className="flex flex-column gap-2">
              <label htmlFor="password" className="text-color font-medium text-md">
                Password
              </label>
              <InputText
                type="password"
                id="password"
                size={45}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            <div className="flex justify-content-between">
              <div>
                <Checkbox
                  checked={rememberMe}
                  inputId="remember_me"
                  name="remember_me"
                  onChange={() => setRememberMe((currentValue) => !currentValue)}
                />
                <label htmlFor="remember_me" className="ml-2 text-sm">
                  Remember me
                </label>
              </div>
              <div className="text-sm cursor-pointer text-teal-700 font-medium">
                Forgot Password?
              </div>
            </div>
            <div className="flex justify-content-center">
              <span className="text-red-500 text-sm">{errorMsg}</span>
            </div>
            <div>
              <Button
                className="w-full border-round-3xl flex  gap-3 justify-content-center"
                raised
                size="small"
                onClick={handleLogin}
              >
                <span className="font-semibold">Sign In</span>
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
export default LoginPage;