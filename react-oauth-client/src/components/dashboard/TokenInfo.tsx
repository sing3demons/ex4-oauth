import React, { useState, useEffect } from 'react';
import { StorageUtils } from '../../utils/oauth';
import { toast } from 'react-hot-toast';

interface TokenInfo {
  accessToken: string | null;
  refreshToken: string | null;
  idToken: string | null;
  isValid: boolean;
  expiresAt: string | null;
  payload: any;
}

export const TokenInfo: React.FC = () => {
  const [tokenInfo, setTokenInfo] = useState<TokenInfo>({
    accessToken: null,
    refreshToken: null,
    idToken: null,
    isValid: false,
    expiresAt: null,
    payload: null
  });
  const [showFullTokens, setShowFullTokens] = useState(false);

  useEffect(() => {
    loadTokenInfo();
  }, []);

  const loadTokenInfo = () => {
    const accessToken = StorageUtils.getAccessToken();
    const refreshToken = StorageUtils.getRefreshToken();
    const idToken = StorageUtils.getIdToken();

    let isValid = false;
    let expiresAt = null;
    let payload = null;

    if (accessToken) {
      try {
        payload = JSON.parse(atob(accessToken.split('.')[1]));
        isValid = payload.exp * 1000 > Date.now();
        expiresAt = new Date(payload.exp * 1000).toLocaleString();
      } catch (error) {
        console.error('Failed to parse token:', error);
      }
    }

    setTokenInfo({
      accessToken,
      refreshToken,
      idToken,
      isValid,
      expiresAt,
      payload
    });
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    toast.success(`${label} copied to clipboard`);
  };

  const formatToken = (token: string) => {
    if (showFullTokens) return token;
    return `${token.substring(0, 30)}...${token.substring(token.length - 30)}`;
  };

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-6 py-5 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-medium text-gray-900">Token Information</h3>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => setShowFullTokens(!showFullTokens)}
              className="text-sm text-blue-600 hover:text-blue-800"
            >
              {showFullTokens ? 'Hide Full Tokens' : 'Show Full Tokens'}
            </button>
            <button
              onClick={loadTokenInfo}
              className="text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700"
            >
              Refresh
            </button>
          </div>
        </div>
      </div>
      
      <div className="px-6 py-5">
        <div className="space-y-6">
          {/* Token Status */}
          <div className="flex items-center space-x-2">
            <span className="text-sm font-medium text-gray-700">Status:</span>
            <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
              tokenInfo.isValid ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
            }`}>
              {tokenInfo.isValid ? '✓ Valid' : '✗ Invalid/Expired'}
            </span>
            {tokenInfo.expiresAt && (
              <span className="text-sm text-gray-500">
                Expires: {tokenInfo.expiresAt}
              </span>
            )}
          </div>

          {/* Access Token */}
          {tokenInfo.accessToken && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Access Token
              </label>
              <div className="bg-gray-50 p-3 rounded-md">
                <div className="flex items-center justify-between mb-2">
                  <code className="text-xs text-gray-600 break-all">
                    {formatToken(tokenInfo.accessToken)}
                  </code>
                  <button
                    onClick={() => copyToClipboard(tokenInfo.accessToken!, 'Access token')}
                    className="ml-2 text-blue-600 hover:text-blue-800 text-sm whitespace-nowrap"
                  >
                    Copy
                  </button>
                </div>
                {tokenInfo.payload && (
                  <details className="text-xs">
                    <summary className="cursor-pointer text-gray-500 hover:text-gray-700">
                      View Payload
                    </summary>
                    <pre className="mt-2 bg-gray-100 p-2 rounded text-xs overflow-auto">
                      {JSON.stringify(tokenInfo.payload, null, 2)}
                    </pre>
                  </details>
                )}
              </div>
            </div>
          )}

          {/* Refresh Token */}
          {tokenInfo.refreshToken && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Refresh Token
              </label>
              <div className="bg-gray-50 p-3 rounded-md">
                <div className="flex items-center justify-between">
                  <code className="text-xs text-gray-600 break-all">
                    {formatToken(tokenInfo.refreshToken)}
                  </code>
                  <button
                    onClick={() => copyToClipboard(tokenInfo.refreshToken!, 'Refresh token')}
                    className="ml-2 text-blue-600 hover:text-blue-800 text-sm whitespace-nowrap"
                  >
                    Copy
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* ID Token */}
          {tokenInfo.idToken && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                ID Token (OIDC)
              </label>
              <div className="bg-gray-50 p-3 rounded-md">
                <div className="flex items-center justify-between">
                  <code className="text-xs text-gray-600 break-all">
                    {formatToken(tokenInfo.idToken)}
                  </code>
                  <button
                    onClick={() => copyToClipboard(tokenInfo.idToken!, 'ID token')}
                    className="ml-2 text-blue-600 hover:text-blue-800 text-sm whitespace-nowrap"
                  >
                    Copy
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};