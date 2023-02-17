/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

#ifndef CONFIG_CURVE_SECP256K1_H
#define CONFIG_CURVE_SECP256K1_H

#include"amcl.h"
#include"config_field_SECP256K1.h"

// ECP stuff

#define CURVETYPE_SECP256K1 WEIERSTRASS
#define PAIRING_FRIENDLY_SECP256K1 NOT
#define CURVE_SECURITY_SECP256K1 128


#if PAIRING_FRIENDLY_SECP256K1 != NOT
//#define USE_GLV_SECP256K1	  /**< Note this method is patented (GLV), so maybe you want to comment this out */
//#define USE_GS_G2_SECP256K1 /**< Well we didn't patent it :) But may be covered by GLV patent :( */
#define USE_GS_GT_SECP256K1 /**< Not patented, so probably safe to always use this */

#define POSITIVEX 0
#define NEGATIVEX 1

#define SEXTIC_TWIST_SECP256K1 .
#define SIGN_OF_X_SECP256K1 .

#define ATE_BITS_SECP256K1

#endif

#if CURVE_SECURITY_SECP256K1 == 128
#define AESKEY_SECP256K1 16 /**< Symmetric Key size - 128 bits */
#define HASH_TYPE_SECP256K1 SHA256  /**< Hash type */
#endif

#if CURVE_SECURITY_SECP256K1 == 192
#define AESKEY_SECP256K1 24 /**< Symmetric Key size - 192 bits */
#define HASH_TYPE_SECP256K1 SHA384  /**< Hash type */
#endif

#if CURVE_SECURITY_SECP256K1 == 256
#define AESKEY_SECP256K1 32 /**< Symmetric Key size - 256 bits */
#define HASH_TYPE_SECP256K1 SHA512  /**< Hash type */
#endif



#endif
