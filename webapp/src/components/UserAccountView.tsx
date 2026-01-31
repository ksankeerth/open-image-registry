import React, { useEffect, useState } from 'react';
import { Dialog } from 'primereact/dialog';
import { Divider } from 'primereact/divider';
import { Button } from 'primereact/button';
import { InputText } from 'primereact/inputtext';
import { Dropdown, DropdownChangeEvent } from 'primereact/dropdown';
import { isValidEmail } from '../utils';
import HttpClient from '../client';
import { useToast } from './ToastComponent';
import { ProgressSpinner } from 'primereact/progressspinner';
import { UserAccountViewDto } from '../api';

export type UserAcccountViewProps = {
  account?: UserAccountViewDto;
  visible: boolean;
  hideCallback: (reloadUsers: boolean) => void;
};

const Roles = ['Developer', 'Maintainer', 'Guest', 'Admin'];

const UserAccountView = (props: UserAcccountViewProps) => {
  const [role, setRole] = useState<string | undefined>(props.account?.role);
  const [email, setEmail] = useState<string | undefined>(props.account?.email);
  const [displayName, setDisplayName] = useState<string | undefined>(props.account?.display_name);

  const [emailValidationMsg, setEmailValidationMsg] = useState<string>('');
  const [roleValidationMsg, setRoleValidationMsg] = useState<string>('');

  const { showSuccess, showError } = useToast();

  const [showProgressView, setShowProgressView] = useState<boolean>(false);

  const roleChangeWarning = 'Changing roles may affect user permissions.';

  // useEffect(() => {
  //   alert("From Account View: " + props.hideCallback)
  // });

  useEffect(() => {
    setDisplayName(props.account?.display_name);
    setRole(props.account?.role);
    setEmail(props.account?.email);
  }, [props.account]);

  const handleUpdateUserInfo = () => {
    if (
      role == props.account?.role &&
      email == props.account?.email &&
      displayName == props.account?.display_name
    ) {
      showSuccess('No changes are detected!');
      return;
    }

    let emailMsg = '';

    if (!email) {
      emailMsg = 'Enter email!';
    } else if (!isValidEmail(email as string)) {
      emailMsg = 'Enter valid email!';
    } else {
      if (!emailValidationMsg.includes('taken')) {
        emailMsg = '';
      } else {
        emailMsg = emailValidationMsg;
      }
    }

    let roleMsg = '';

    if (!role) {
      roleMsg = 'Select a role!';
      setRoleValidationMsg('Select a role!');
    } else {
      setRoleValidationMsg('');
    }

    setEmailValidationMsg(emailMsg);
    setRoleValidationMsg(roleMsg);

    if (emailMsg == '' && roleMsg == '') {
      setShowProgressView(true);
      if (email != props.account?.email) {
        validateUsernameEmailAPICall(updateUserAccountAPICall);
      } else {
        updateUserAccountAPICall();
      }
    }
  };

  const handleDeleteUser = () => {
    HttpClient.getInstance('http://localhost:8000/api/v1')
      .deleteUser(props.account?.id as string)
      .then((data) => {
        if (data.error) {
          showError(data.error);
          setTimeout(() => {
            setShowProgressView(false);
          }, 150);
          return;
        }

        setTimeout(() => {
          setShowProgressView(false);
        }, 200);
        showSuccess('Successfully deleted user account.');
        props.hideCallback(true);
      })
      .catch((err) => {
        console.log(err);
        setTimeout(() => {
          setShowProgressView(false);
        }, 200);
        showError(err);
        return;
      });
  };

  const validateUsernameEmailAPICall = (successFn: () => void) => {
    HttpClient.getInstance('http://localhost:8000/api/v1')
      .valiateUser({
        username: '',
        email: email as string,
      })
      .then((data) => {
        if (data.error) {
          showError(data.error);
          setTimeout(() => {
            setShowProgressView(false);
          }, 200);
          return;
        }
        if (!data.email_available) {
          setEmailValidationMsg('Email is already taken!');
          setTimeout(() => {
            setShowProgressView(false);
          }, 200);
        } else {
          successFn();
        }
      })
      .catch((err) => {
        showError('Unexpected error occurred!');
        setTimeout(() => {
          setShowProgressView(false);
        }, 150);
      });
  };

  const updateUserAccountAPICall = () => {
    HttpClient.getInstance('http://localhost:8000/api/v1')
      .updateUserAccount(
        {
          email: email as string,
          display_name: displayName as string,
          role: role as string,
        },
        props.account?.id as string
      )
      .then((data) => {
        if (data.error) {
          showError(data.error);
          setTimeout(() => {
            setShowProgressView(false);
          }, 150);
          return;
        }
        setTimeout(() => {
          setShowProgressView(false);
        }, 200);
        showSuccess('Successfully update user account.');
        props.hideCallback(true);
      })
      .catch((err) => {
        console.log(err);
        setTimeout(() => {
          setShowProgressView(false);
        }, 200);
        showError(err);
        return;
      });
  };

  const handleRoleChange = (e: DropdownChangeEvent) => {
    setRole(e.target.value);
    if (e.target.value !== props.account?.role) {
    }
  };

  return (
    <React.Fragment>
      <Dialog
        visible={props.visible}
        onHide={() => props.hideCallback(false)}
        className="w-4 p-0 m-0"
        modal
        content={({ hide }) => {
          return (
            <div className="flex flex-column  p-0 m-0 bg-white border-round-lg">
              {/* Header */}
              <div
                className="flex-grow-0  border-round-top-lg 
             flex flex-row  align-items-center justify-content-between gap-2 p-3 pb-0 "
              >
                <div className="font-medium text-lg text-color-secondary">User Account</div>

                <div>
                  <span
                    className="pi pi-times text-sm  cursor-pointer"
                    onClick={(e) => hide(e)}
                  ></span>
                </div>
              </div>
              <div className="flex flex-row justify-content-start align-items-center text-sm gap-2 p-3 pt-0">
                <div>@{props.account?.username}</div>
              </div>
              <Divider className="m-0 p-0" />

              {/* Content */}
              <div className="flex-grow-1 flex flex-column gap-2 p-4">
                <div className=" border-round-lg flex flex-column">
                  <div className="p-0  grid">
                    <div className="col-6 flex align-items-center text-xs required">
                      <i className="pi pi-envelope text-xs "></i>
                      &nbsp;&nbsp;&nbsp; Email Address
                    </div>
                    <div className="col-6 flex text-xs justify-content-end align-items-center pb-0 mb-0">
                      {emailValidationMsg && (
                        <span className="text-red-300">{emailValidationMsg}</span>
                      )}
                    </div>

                    <div className="col-12">
                      <InputText
                        value={email}
                        className="border-1 text-xs"
                        onChange={(e) => setEmail(e.target.value)}
                      />
                    </div>

                    <div className="col-5 text-xs flex align-items-center">
                      <i className="pi pi-user text-xs"></i>
                      &nbsp;&nbsp;&nbsp; Display Name
                    </div>
                    <div className="col-2 text-xs"></div>
                    <div className="col-5 text-xs flex align-items-center required">
                      <i className="pi pi-shield text-xs"></i>
                      &nbsp;&nbsp;&nbsp; Role
                    </div>
                    <div className="col-5">
                      <InputText
                        value={displayName}
                        className="border-1 text-xs"
                        onChange={(e) => setDisplayName(e.target.value)}
                      />
                    </div>
                    <div className="col-2"></div>
                    <div className="col-5">
                      <Dropdown
                        pt={{
                          input: {
                            className: 'text-xs',
                          },
                        }}
                        className="w-full border-1 p-0 text-xs"
                        panelClassName="text-xs animate-fadein"
                        options={Roles}
                        value={role}
                        onChange={handleRoleChange}
                      />
                    </div>
                    {roleValidationMsg && <div className="col-12  text-red-300 text-xs "></div>}

                    {role != props.account?.role && (
                      <div className="col-12 text-xs flex flex-row justify-content-center">
                        <span className="text-red-300">{roleChangeWarning}</span>
                      </div>
                    )}

                    <div className="col-12 flex flex-row justify-content-end">
                      <Button
                        size="small"
                        outlined
                        className="border-round-3xl border-1 text-xs"
                        onClick={handleUpdateUserInfo}
                      >
                        Update
                      </Button>
                    </div>
                  </div>
                </div>
              </div>

              <div className="flex flex-column gap-3 ml-4 mr-4 mb-3 p-3 border-1 border-red-200 border-round-lg">
                <div className="text-sm">
                  <span className="pi pi-exclamation-triangle text-red-600"></span>
                  &nbsp;&nbsp;&nbsp;
                  <span className="text-red-800">Danger Zone</span>
                </div>
                <div className="text-red-600 text-xs">
                  Deleting this user will permanently remove all associated data. This action cannot
                  be undone.
                </div>
                <Button
                  size="small"
                  severity="danger"
                  className="flex justify-content-center border-round-3xl border-1"
                  onClick={handleDeleteUser}
                >
                  <span className="pi pi-trash"></span>
                  &nbsp;&nbsp;&nbsp;
                  <span>Delete User Account</span>
                </Button>
              </div>
              <div className="flex flex-row justify-content-end">
                <div></div>
              </div>
            </div>
          );
        }}
      ></Dialog>
      {showProgressView && (
        <div
          className="fixed top-0 left-50 bottom-0 h-full w-full surface-50 opacity-70 flex align-items-center justify-content-center"
          style={{ zIndex: 1000 }}
        >
          <div className="flex flex-column align-items-center">
            <ProgressSpinner style={{ width: '50px', height: '50px' }} />
          </div>
        </div>
      )}
    </React.Fragment>
  );
};

export default UserAccountView;
