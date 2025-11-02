import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import accountsetup from "./../assets/account-setup1.png";
import { Button } from "primereact/button";
import { InputText } from "primereact/inputtext";
import LogoComponent from "../components/LogoComponent";
import "./login.css";
import HttpClient from "../client";
import { useToast } from "../components/ToastComponent";
import { validatePassword } from "../utils";

const NewAccountSetupPage = () => {
  // Separate validation states for better control
  const [generalError, setGeneralError] = useState<string>("");
  const [displayNameError, setDisplayNameError] = useState<string>("");
  const [passwordError, setPasswordError] = useState<string>("");
  const [passwordMatchError, setPasswordMatchError] = useState<string>("");

  const [timerCountDown, setTimerCountDown] = useState<number>(30);
  const [displayName, setDisplayName] = useState<string>("");
  const [username, setUsername] = useState<string>("");
  const [role, setRole] = useState<string>("");
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [password1, setPassword1] = useState<string>("");
  const [userId, setUserId] = useState<string>("");
  const [completed, setCompleted] = useState<boolean>(false);
  const [showRedirectMsg, setShowRedirectMsg] = useState<boolean>(false);
  const [touched, setTouched] = useState({
    displayName: false,
    password: false,
    password1: false,
  });

  const navigate = useNavigate();
  const { showSuccess, showError } = useToast();
  const { uuid } = useParams();

  // Load account setup info
  useEffect(() => {
    if (!uuid) {
      return;
    }

    HttpClient.getInstance()
      .getAccountSetupInfo(uuid)
      .then((data) => {
        if (data.error_message) {
          setGeneralError(data.error_message);
          setShowRedirectMsg(true);
          setTimerCountDown(30);

          const interval = setInterval(() => {
            setTimerCountDown((prev) => {
              if (prev === 0) {
                clearInterval(interval);
                return 0;
              }
              return prev - 1;
            });
          }, 1000);

          return;
        }
        setUsername(data.username);
        setDisplayName(data.display_name);
        setEmail(data.email);
        setRole(data.role);
        setUserId(data.user_id);
      })
      .catch((err) => {
        console.log(err);
        setGeneralError("Failed to load account information");
      });
  }, [uuid]);

  // Validate display name (only when touched)
  useEffect(() => {
    if (!touched.displayName) {
      return;
    }

    if (!displayName || displayName.trim() === "") {
      setDisplayNameError("Display name is required");
    } else {
      setDisplayNameError("");
    }
  }, [displayName, touched.displayName]);

  // Validate password strength
  useEffect(() => {
    if (!touched.password || !password) {
      setPasswordError("");
      return;
    }

    const res = validatePassword(password);
    if (!res.isValid) {
      setPasswordError(res.msg);
    } else {
      setPasswordError("");
    }
  }, [password, touched.password]);

  // Validate password match
  useEffect(() => {
    if (!touched.password1 || !password1) {
      setPasswordMatchError("");
      return;
    }

    if (password !== password1) {
      setPasswordMatchError("Passwords do not match");
    } else {
      setPasswordMatchError("");
    }
  }, [password, password1, touched.password1]);

  // Check if form is valid
  const isFormValid = () => {
    return (
      displayName.trim() !== "" &&
      password !== "" &&
      password1 !== "" &&
      password === password1 &&
      validatePassword(password).isValid &&
      !generalError
    );
  };

  const handleUserAccountComplete = () => {
    // Mark all fields as touched for validation
    setTouched({
      displayName: true,
      password: true,
      password1: true,
    });

    // Validate all fields
    if (!displayName || displayName.trim() === "") {
      setDisplayNameError("Display name is required");
      return;
    }

    if (!password || !password1) {
      setGeneralError("Please enter both password fields");
      return;
    }

    if (password !== password1) {
      setPasswordMatchError("Passwords do not match");
      return;
    }

    const pwValidation = validatePassword(password);
    if (!pwValidation.isValid) {
      setPasswordError(pwValidation.msg);
      return;
    }

    setShowRedirectMsg(false);

    HttpClient.getInstance()
      .completeAccountSetup({
        user_id: userId,
        username: username,
        display_name: displayName,
        password: password,
        uuid: uuid as string,
      })
      .then((data: { error_message?: string }) => {
        if (data.error_message) {
          showError(data.error_message);
          setGeneralError(data.error_message);
        } else {
          showSuccess("Successfully completed account setup!");
          setCompleted(true);
          setShowRedirectMsg(true);
          setTimerCountDown(30);

          const interval = setInterval(() => {
            setTimerCountDown((prev) => {
              if (prev === 0) {
                clearInterval(interval);
                navigate("/login");
                return 0;
              }
              return prev - 1;
            });
          }, 1000);
        }
      })
      .catch((err) => {
        setGeneralError("An error occurred while completing setup");
        showError("An error occurred while completing setup");
      });
  };

  return (
    <div className="flex flex-row min-h-screen max-h-screen">
      <div className="w-6 flex-column  login-left-container backdrop-blur-3xl!">
        {/* ... animation container stays the same ... */}
        <div
          className="animation-container w-full backdrop-blur-2xl!"
          style={{ maxHeight: "60vh" }}
        >
          <div className="sky">
            <div className="cloud cloud1"></div>
            <div className="cloud cloud2"></div>
          </div>

          <div className="seagull seagull1"></div>
          <div className="seagull seagull2"></div>

          <div className="water">
            <div className="wave"></div>
          </div>

          <div className="dock"></div>

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
        <div
          className="flex flex-row justify-content-end"
          style={{ maxHeight: "40vh" }}
        >
          <img
            src={accountsetup}
            style={{ maxHeight: "40vh" }}
            className="flex-grow-1"
          />
        </div>
      </div>

      <div className="w-6 flex align-items-center justify-content-center relative">
        <div className="flex flex-column" style={{ maxWidth: "400px" }}>
          <div className="flex flex-row justify-content-center gap-2 pb-2">
            <span className="text-green-300" style={{ fontSize: 20 }}>
              Let's
            </span>
            <span style={{ fontSize: 20, color: "#007700" }}>COMPLETE</span>
            <span className="text-green-300" style={{ fontSize: 20 }}>
              Onboarding!
            </span>
          </div>
          <LogoComponent showNameInOneLine={true} />

          {/* General messages */}
          <div className="w-full flex flex-column align-items-center justify-content-center text-color font-medium text-sm pb-3 pt-4">
            {generalError && (<span className="text-red-300 text-center">{generalError}</span>)}
            {!generalError && !completed && (
              <span className="text-600">
                Check details and enter valid password!
              </span>
            )}
            {showRedirectMsg && (
              <span className="text-600 mt-2">
                Redirecting to Sign In page in {timerCountDown} seconds.
              </span>
            )}
          </div>

          <div className="flex flex-column gap-3">
            {/* Email */}
            <div className="flex flex-column gap-2">
              <label htmlFor="email" className="text-color font-medium text-md">
                Email
              </label>
              <InputText id="email" disabled size={45} value={email} />
            </div>

            {/* Role */}
            <div className="flex flex-column gap-2">
              <label htmlFor="role" className="text-color font-medium text-md">
                Role
              </label>
              <InputText id="role" disabled size={45} value={role} />
            </div>

            {/* Username */}
            <div className="flex flex-column gap-2">
              <label
                htmlFor="username"
                className="text-color font-medium text-md"
              >
                Username
              </label>
              <InputText id="username" size={45} disabled value={username} />
            </div>

            {/* Display Name */}
            <div className="flex flex-column gap-2">
              <label
                htmlFor="displayname"
                className="text-color font-medium text-md required"
              >
                Display Name
              </label>
              <InputText
                id="displayname"
                size={45}
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                onBlur={() => setTouched({ ...touched, displayName: true })}
                className={displayNameError ? "p-invalid" : ""}
              />
              {displayNameError && (
                <small className="p-error">{displayNameError}</small>
              )}
            </div>

            {/* New Password */}
            <div className="flex flex-column gap-2">
              <div className="flex flex-row align-items-center gap-2">
                <label
                  htmlFor="password"
                  className="text-color font-medium text-md required"
                >
                  New Password
                </label>
                <span
                  className="pi pi-question-circle text-xs cursor-pointer text-700"
                  title="Password must be at least 8 characters long and contain uppercase, lowercase, numbers, and special characters"
                ></span>
              </div>
              <InputText
                type="password"
                id="password"
                size={45}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onBlur={() => setTouched({ ...touched, password: true })}
                className={passwordError ? "p-invalid" : ""}
              />
              {passwordError && (
                <small className="p-error">{passwordError}</small>
              )}
              {!passwordError && touched.password && password && (
                <small className="text-green-600">
                  ✓ Password meets requirements
                </small>
              )}
            </div>

            {/* Re-enter Password */}
            <div className="flex flex-column gap-2">
              <label
                htmlFor="password1"
                className="text-color font-medium text-md required"
              >
                Re-enter Password
              </label>
              <InputText
                type="password"
                id="password1"
                size={45}
                value={password1}
                onChange={(e) => setPassword1(e.target.value)}
                onBlur={() => setTouched({ ...touched, password1: true })}
                className={passwordMatchError ? "p-invalid" : ""}
              />
              {passwordMatchError && (
                <small className="p-error">{passwordMatchError}</small>
              )}
              {!passwordMatchError &&
                touched.password1 &&
                password1 &&
                password === password1 && (
                  <small className="text-green-600">✓ Passwords match</small>
                )}
            </div>

            {/* Submit Button */}
            <div>
              <Button
                className="w-full flex justify-content-center"
                raised
                size="small"
                disabled={!isFormValid()}
                onClick={handleUserAccountComplete}
                tooltip={
                  !isFormValid()
                    ? "Please fill all required fields correctly"
                    : ""
                }
                tooltipOptions={{ position: "top" }}
              >
                Complete
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default NewAccountSetupPage;