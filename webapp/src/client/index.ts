import { TableFilterSearchPaginationSortState, UserAccountInfo } from "../types/app_types";
import {
  AccountSetupCompleteRequest,
  AuthLoginRequest,
  AuthLoginResponse,
  CreateUserAccountRequest,
  CreateUserAccountResponse,
  ListUpstreamRegistriesResponse,
  ListUsersResponse,
  PostUpstreamRequestBody,
  PostUpstreamResponseBody,
  UpdateUserAccountRequest,
  UpdateUserAccountResponse,
  UserAccountSetupInfoResponse,
  UsernameEmailValidationRequest,
  UsernameEmailValidationResponse,
} from "../types/request_response";
import { parseDates } from "../utils/dateParser";

export default class HttpClient {
  private static instance: HttpClient;
  private baseUrl: string;

  private constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  public static getInstance(baseUrl: string = "http://localhost:8000/api/v1"): HttpClient {
    if (!HttpClient.instance) {
      HttpClient.instance = new HttpClient(baseUrl);
    }
    return HttpClient.instance;
  }

  public async login(request: AuthLoginRequest): Promise<AuthLoginResponse> {
    try {
      const response = await fetch(`${this.baseUrl}/auth/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      const body = await response.json();

      if (
        response.status == 200 ||
        response.status == 500 ||
        response.status == 403
      ) {
        return body as AuthLoginResponse;
      } else {
        return {
          success: false,
          error_message: "Unexpected error occurred. Please try again!",
        } as AuthLoginResponse;
      }
    } catch (err) {
      console.log(err);
      return {
        success: false,
        error_message: "Unexpected error occurred. Please try again!",
      } as AuthLoginResponse;
    }
  }

  public async createUpstream(
    request: PostUpstreamRequestBody
  ): Promise<PostUpstreamResponseBody> {
    try {
      const response = await fetch(`${this.baseUrl}/upstreams`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      const body = await response.json();

      if (response.status == 201) {
        return {
          ...(body as { reg_id: string; reg_name: string }),
        };
      }
      if (response.status == 409) {
        return {
          error: "Port or name may conflict with existing registry.",
        };
      }
      return {
        error: "Error occured when creating Upstream OCI Registry",
      };
    } catch (err) {
      console.log(err);
      return {
        error: "Error occured when creating Upstream OCI Registry",
      };
    }
  }

  public async getUpstreamRegisteries(): Promise<
    ListUpstreamRegistriesResponse | { error: string }
  > {
    try {
      const response = await fetch(`${this.baseUrl}/upstreams`, {
        method: "GET",
        headers: {
          Accept: "application/json",
        },
      });

      const resBody = await response.json();

      if (response.ok) {
        return resBody as ListUpstreamRegistriesResponse;
      }
      return {
        error: "Error occured when retriving upstream registeries",
      };
    } catch (error) {
      console.log(error);
      return {
        error: "Error occured when retriving upstream registeries",
      };
    }
  }

  public async createUserAccount(
    request: CreateUserAccountRequest
  ): Promise<CreateUserAccountResponse> {
    try {
      const response = await fetch(`${this.baseUrl}/users`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      const body = await response.json();

      if (response.status == 201) {
        return body as CreateUserAccountResponse;
      } else if (response.status == 400) {
        return {
          error: "Provide valid values for Email and Username!",
        } as CreateUserAccountResponse;
      } else if (response.status == 409) {
        return {
          error: "Email or Username are not available!",
        } as CreateUserAccountResponse;
      } else {
        return {
          error: "Unexpected error occurred. Please try again!",
        } as CreateUserAccountResponse;
      }
    } catch (err) {
      console.log(err);
      return {
        error: "Error occurred when creating user account",
      } as CreateUserAccountResponse;
    }
  }

  public async valiateUser(
    request: UsernameEmailValidationRequest
  ): Promise<UsernameEmailValidationResponse> {
    try {
      const response = await fetch(`${this.baseUrl}/users/validate`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      const body = await response.json();

      if (response.status == 200) {
        return body as UsernameEmailValidationResponse;
      }
      if (response.status == 400) {
        console.log("invalid request body");
      }
      return {
        error: "Unexpected error occurred!",
      } as UsernameEmailValidationResponse;
    } catch (err) {
      console.log(err);
      return {
        error: "Unexpected error occurred!",
      } as UsernameEmailValidationResponse;
    }
  }

  public async updateUserAccount(
    request: UpdateUserAccountRequest, userId: string
  ): Promise<UpdateUserAccountResponse> {
    try {
      const response = await fetch(`${this.baseUrl}/users/${userId}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      if (response.status == 200) {
        return {};
      }
      if (response.status == 400) {
        console.log("invalid request body");
      }
      return {
        error: "Unexpected error occured! Please try again.",
      };
    } catch (err) {
      console.log(err);
      return {
        error: "Unexpected error occured! Please try again.",
      };
    }
  }

  private buildTableQueryParams<T>(state: TableFilterSearchPaginationSortState<T>): URLSearchParams {
    const params = new URLSearchParams();

    params.append('page', state.pagination.page.toString());
    params.append('limit', state.pagination.limit.toString());

    if (state.sort && state.sort.key) {
      const sortFieldName = state.sort.key;
      params.append('sort_by', sortFieldName as string);

      // Convert order: -1 -> desc, 1 -> asc, 0 -> skip
      if (state.sort.order === 1) {
        params.append('order', 'asc');
      } else if (state.sort.order === -1) {
        params.append('order', 'desc');
      }
    }

    if (state.search_value && state.search_value.trim()) {
      params.append('search', state.search_value.trim());
    }

    if (state.filters && state.filters.length > 0) {
      for (const filter of state.filters) {
        const filterParamName = filter.key;
        for (const value of filter.values) {
          params.append(filterParamName as string, String(value));
        }
      }
    }

    return params;
  }

  public async getUsersList(params: TableFilterSearchPaginationSortState<UserAccountInfo>): Promise<ListUsersResponse> {
    try {
      const queryParams = this.buildTableQueryParams<UserAccountInfo>(params)
      const response = await fetch(`${this.baseUrl}/users?${queryParams.toString()}`, {
        method: "GET",
      });

      const body = await response.json();

      if (response.status == 200) {
        return (parseDates(body)) as ListUsersResponse;
      }
      //TODO: handle 403/401 later
      return {
        error: "Unexpected error occurred!"
      } as ListUsersResponse;

    } catch (err) {
      console.log(err);
      return {
        error: "Unexpected error occurred!"
      } as ListUsersResponse;
    }
  }

  public async deleteUser(userId: string): Promise<{ error?: string }> {
    try {
      const response = await fetch(`${this.baseUrl}/users/${userId}`, {
        method: "DELETE",
        headers: {
        }
      });

      if (response.status == 200) {
        return {};
      }
      if (response.status == 400) {
        console.log("invalid request body");
      }
      return {
        error: "Unexpected error occured! Please try again.",
      };
    } catch (err) {
      console.log(err);
      return {
        error: "Unexpected error occured! Please try again.",
      };
    }
  }

  public async getAccountSetupInfo(uuid: string): Promise<UserAccountSetupInfoResponse> {
    try {
      const response = await fetch(`${this.baseUrl}/users/account-setup/${uuid}`, {
        method: "GET",
        headers: {
        }
      });

      const body = await response.json();

      if (response.status == 404) {
        return body as UserAccountSetupInfoResponse
      }

      if (response.status == 200) {
        return body as UserAccountSetupInfoResponse;
      }
      if (response.status == 500) {
        return { error_message: "Unexpected error occurred! Please contact administrator." } as UserAccountSetupInfoResponse
      }
      return {
        error_message: "Unexpected error occured! Please try again.",
      } as UserAccountSetupInfoResponse;
    } catch (err) {
      console.log(err);
      return {
        error_message: "Unexpected error occured! Please try again.",
      } as UserAccountSetupInfoResponse;
    }
  }

  public async completeAccountSetup(
    request: AccountSetupCompleteRequest
  ): Promise<{ error_message?: string }> {
    try {
      const response = await fetch(`${this.baseUrl}/users/account-setup/${request.uuid}/complete`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });


      if (response.status == 200) {
        return { error_message: "" };
      }
      return { error_message: "Unexpected error occurred! Please try again." };

    } catch (err) {
      console.log(err);
      return { error_message: "Unexpected error occurred! Please try again." };
    }
  }
}