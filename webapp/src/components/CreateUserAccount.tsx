import { Button } from 'primereact/button';
import { Dialog } from 'primereact/dialog';
import { Dropdown } from 'primereact/dropdown';
import { InputText } from 'primereact/inputtext';
import React, { useState } from 'react';
import { isValidEmail, validateUsernameWithError } from '../utils';
import { useToast } from './ToastComponent';
import { classNames } from 'primereact/utils';
import { Divider } from 'primereact/divider';
import { useLoader } from './loader';
import { postUsers, postUsersValidate } from '../api';

const Roles = ['Developer', 'Maintainer', 'Guest', 'Admin'];

export type CreateUserAccountDialogProps = {
  visible: boolean;
  hideCallback: (reloadUsers: boolean) => void;
};

const CreateUserAccountDialog = (props: CreateUserAccountDialogProps) => {
  const [email, setEmail] = useState<string>('');
  const [displayName, setDisplayName] = useState<string>('');
  const [username, setUsername] = useState<string>('');
  const [role, setRole] = useState<'Admin' | 'Developer' | 'Guest' | 'Maintainer'>('Guest');

  const [emailValidationMsg, setEmailValidationMsg] = useState<string>('');
  const [roleValidationMsg, setRoleValidationMsg] = useState<string>('');
  const [usernameValidationMsg, setUsernameValidationMsg] = useState<string>('');

  const { showSuccess, showError } = useToast();
  const { showLoading, hideLoading } = useLoader();

  const resetFields = () => {
    setEmailValidationMsg('');
    setRoleValidationMsg('');
    setUsernameValidationMsg('');
    setUsername('');
    setEmail('');
    setRole('Guest');
    setDisplayName('');
  };

  const handleButtonClick = () => {
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

    let usernameMsg = '';
    const res = validateUsernameWithError(username as string);
    if (!username) {
      usernameMsg = 'Enter a username. Users can update it during account setup.';
    } else if (!res.isValid) {
      usernameMsg = res.error as string;
    } else {
      if (!usernameValidationMsg.includes('taken')) {
        usernameMsg = '';
      } else {
        usernameMsg = usernameValidationMsg;
      }
    }

    setEmailValidationMsg(emailMsg);
    setUsernameValidationMsg(usernameMsg);
    setRoleValidationMsg(roleMsg);

    if (emailMsg == '' && roleMsg == '' && usernameMsg == '') {
      validateUsernameEmailAPICall(createUserAccountAPICall);
    }
  };

  const validateUsernameEmailAPICall = async (successFn: () => void) => {
    showLoading('Checking details ...');
    const { data, error } = await postUsersValidate({
      body: {
        username: username,
        email: email,
      },
    });

    hideLoading();
    if (error) {
      showError(error.error_message);
      return;
    }

    if (!data.email_available) {
      setEmailValidationMsg('Email is already taken!');
      hideLoading();
    } else if (!data.username_available) {
      setUsernameValidationMsg('Username is already taken!');
      hideLoading();
    } else {
      successFn();
    }
  };

  const createUserAccountAPICall = async () => {
    showLoading('Creating user account');

    const { data, error } = await postUsers({
      body: {
        username,
        email,
        role,
        display_name: displayName,
      },
    });
    hideLoading();
    if (error) {
      showError(error.error_message);
    }

    if (data?.user_id) {
      showSuccess('Successfully created user account & sent invitation to ' + email);
      resetFields();
      props.hideCallback(true);
    }
  };

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    setEmailValidationMsg('');
  };

  const handleUsernameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUsername(e.target.value);
    setUsernameValidationMsg('');
  };

  return (
    <React.Fragment>
      <Dialog
        visible={props.visible}
        onHide={() => props.hideCallback(false)}
        modal
        dismissableMask
        showHeader={false}
        pt={{
          root: {
            className: 'border-none shadow-none',
          },
          mask: {
            style: {
              backdropFilter: 'blur(8px)',
              backgroundColor: 'rgba(0, 0, 0, 0.3)',
            },
          },
          content: {
            className: 'p-0 m-0 border-round-2xl overflow-hidden',
            style: {
              width: '420px',
              animation: 'dialogSlideIn 0.3s ease-out',
            },
          },
        }}
      >
        <div className="bg-white">
          {/* Minimal Header */}
          <div className="px-4 pt-5 pb-2">
            <div className="flex align-items-start justify-content-between mb-1">
              <h2 className="m-0 text-2xl font-semibold text-gray-900">Create User</h2>
              <Button
                icon="pi pi-times"
                onClick={() => props.hideCallback(false)}
                text
                rounded
                className="w-2rem h-2rem -mt-1 -mr-2"
                pt={{
                  root: {
                    className: 'text-gray-400 hover:text-gray-600 hover:bg-gray-50',
                  },
                }}
              />
            </div>
            <p className="m-0 text-sm text-gray-500">Send an invitation to join registry</p>
          </div>
          <Divider layout="horizontal" className="pt-0 mt-0" />
          {/* Form Fields */}
          <div className="px-4 pb-4">
            <div className="flex flex-column" style={{ gap: '1.25rem' }}>
              {/* Email */}
              <div className="flex flex-column" style={{ gap: '0.5rem' }}>
                <label htmlFor="email" className="text-sm font-medium text-gray-700">
                  Email
                </label>
                <InputText
                  id="email"
                  value={email}
                  onChange={handleEmailChange}
                  placeholder="user@example.com"
                  className={classNames('w-full border-round-3xl', {
                    'p-invalid': emailValidationMsg,
                  })}
                  pt={{
                    root: {
                      className:
                        'text-sm p-3 py-2 border-1 border-gray-200 hover:border-gray-300 focus:border-teal-500',
                      style: {
                        transition: 'all 0.2s ease',
                        boxShadow: emailValidationMsg ? '0 0 0 1px #ef4444' : 'none',
                      },
                    },
                  }}
                />
                {emailValidationMsg && (
                  <small className="text-red-500 text-xs">{emailValidationMsg}</small>
                )}
              </div>

              {/* Role */}
              <div className="flex flex-column" style={{ gap: '0.5rem' }}>
                <label htmlFor="role" className="text-sm font-medium text-gray-700">
                  Role
                </label>
                <Dropdown
                  id="role"
                  value={role}
                  onChange={(e) => setRole(e.value)}
                  options={Roles}
                  placeholder="Select role"
                  className={classNames('w-full border-round-3xl ', {
                    'p-invalid': roleValidationMsg,
                  })}
                  pt={{
                    root: {
                      className: 'border-1  border-gray-200 hover:border-gray-300',
                      style: {
                        transition: 'all 0.2s ease',
                        boxShadow: roleValidationMsg ? '0 0 0 1px #ef4444' : 'none',
                      },
                    },
                    input: {
                      className: 'text-sm ',
                    },
                    item: {
                      className: 'text-sm px-3 py-2',
                    },
                    panel: {
                      className: 'border-1 border-gray-200 shadow-lg',
                    },
                  }}
                />
                {roleValidationMsg && (
                  <small className="text-red-500 text-xs">{roleValidationMsg}</small>
                )}
              </div>

              {/* Username & Display Name */}
              <div className="grid" style={{ margin: 0 }}>
                <div className="col-6 pl-0">
                  <div className="flex flex-column" style={{ gap: '0.5rem' }}>
                    <label htmlFor="username" className="text-sm font-medium text-gray-700">
                      Username
                    </label>
                    <InputText
                      id="username"
                      value={username}
                      onChange={handleUsernameChange}
                      placeholder="johndoe"
                      className={classNames('w-full border-round-3xl', {
                        'p-invalid': usernameValidationMsg,
                      })}
                      pt={{
                        root: {
                          className:
                            'text-sm p-3 py-2 border-1 border-gray-200 hover:border-gray-300 focus:border-teal-500',
                          style: {
                            transition: 'all 0.2s ease',
                            boxShadow: usernameValidationMsg ? '0 0 0 1px #ef4444' : 'none',
                          },
                        },
                      }}
                    />
                    {usernameValidationMsg && (
                      <small className="text-red-500 text-xs" style={{ lineHeight: '1.3' }}>
                        {usernameValidationMsg}
                      </small>
                    )}
                  </div>
                </div>

                <div className="col-6 pr-0">
                  <div className="flex flex-column" style={{ gap: '0.5rem' }}>
                    <label htmlFor="displayName" className="text-sm font-medium text-gray-700">
                      Display Name
                    </label>
                    <InputText
                      id="displayName"
                      value={displayName}
                      onChange={(e) => setDisplayName(e.target.value)}
                      placeholder="John Doe"
                      className="w-full border-round-3xl"
                      pt={{
                        root: {
                          className:
                            'text-sm p-3 py-2 border-1 border-gray-200 hover:border-gray-300 focus:border-teal-500',
                          style: {
                            transition: 'all 0.2s ease',
                          },
                        },
                      }}
                    />
                    <small className="text-xs text-gray-400 pl-1">Optional</small>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Footer with Actions */}
          <div className="px-4 py-3 bg-gray-50 flex justify-content-end" style={{ gap: '0.75rem' }}>
            <Button
              label="Cancel"
              onClick={() => props.hideCallback(false)}
              text
              className="text-sm px-4 py-2 border-round-3xl"
              pt={{
                root: {
                  className: 'text-gray-600 hover:bg-gray-100',
                  style: {
                    transition: 'all 0.2s ease',
                  },
                },
              }}
            />
            <Button
              label="Invite"
              icon="pi pi-send"
              iconPos="left"
              onClick={handleButtonClick}
              className="text-sm px-4 py-2 border-round-3xl"
              pt={{
                root: {
                  className: 'hover:shadow-2',
                  style: {
                    transition: 'all 0.2s ease',
                  },
                },
              }}
            />
          </div>
        </div>
      </Dialog>
    </React.Fragment>
  );
};

export default CreateUserAccountDialog;
