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

/**
 * @file ecp_SECP256K1.h
 * @author Mike Scott
 * @brief ECP Header File
 *
 */

#ifndef ECP_SECP256K1_H
#define ECP_SECP256K1_H

#include "fp_SECP256K1.h"
#include "config_curve_SECP256K1.h"

/* Curve Params - see rom_zzz.c */
extern const int CURVE_A_SECP256K1;         /**< Elliptic curve A parameter */
extern const int CURVE_Cof_I_SECP256K1;     /**< Elliptic curve cofactor */
extern const int CURVE_B_I_SECP256K1;       /**< Elliptic curve B_i parameter */
extern const BIG_256_56 CURVE_B_SECP256K1;     /**< Elliptic curve B parameter */
extern const BIG_256_56 CURVE_Order_SECP256K1; /**< Elliptic curve group order */
extern const BIG_256_56 CURVE_Cof_SECP256K1;   /**< Elliptic curve cofactor */

/* Generator point on G1 */
extern const BIG_256_56 CURVE_Gx_SECP256K1; /**< x-coordinate of generator point in group G1  */
extern const BIG_256_56 CURVE_Gy_SECP256K1; /**< y-coordinate of generator point in group G1  */


/* For Pairings only */

/* Generator point on G2 */
extern const BIG_256_56 CURVE_Pxa_SECP256K1; /**< real part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxb_SECP256K1; /**< imaginary part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pya_SECP256K1; /**< real part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pyb_SECP256K1; /**< imaginary part of y-coordinate of generator point in group G2 */


/*** needed for BLS24 curves ***/

extern const BIG_256_56 CURVE_Pxaa_SECP256K1; /**< real part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxab_SECP256K1; /**< imaginary part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxba_SECP256K1; /**< real part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxbb_SECP256K1; /**< imaginary part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pyaa_SECP256K1; /**< real part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pyab_SECP256K1; /**< imaginary part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pyba_SECP256K1; /**< real part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pybb_SECP256K1; /**< imaginary part of y-coordinate of generator point in group G2 */

/*** needed for BLS48 curves ***/

extern const BIG_256_56 CURVE_Pxaaa_SECP256K1; /**< real part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxaab_SECP256K1; /**< imaginary part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxaba_SECP256K1; /**< real part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxabb_SECP256K1; /**< imaginary part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxbaa_SECP256K1; /**< real part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxbab_SECP256K1; /**< imaginary part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxbba_SECP256K1; /**< real part of x-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pxbbb_SECP256K1; /**< imaginary part of x-coordinate of generator point in group G2 */

extern const BIG_256_56 CURVE_Pyaaa_SECP256K1; /**< real part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pyaab_SECP256K1; /**< imaginary part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pyaba_SECP256K1; /**< real part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pyabb_SECP256K1; /**< imaginary part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pybaa_SECP256K1; /**< real part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pybab_SECP256K1; /**< imaginary part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pybba_SECP256K1; /**< real part of y-coordinate of generator point in group G2 */
extern const BIG_256_56 CURVE_Pybbb_SECP256K1; /**< imaginary part of y-coordinate of generator point in group G2 */


extern const BIG_256_56 CURVE_Bnx_SECP256K1; /**< BN curve x parameter */

extern const BIG_256_56 CURVE_Cru_SECP256K1; /**< BN curve Cube Root of Unity */

extern const BIG_256_56 Fra_SECP256K1; /**< real part of BN curve Frobenius Constant */
extern const BIG_256_56 Frb_SECP256K1; /**< imaginary part of BN curve Frobenius Constant */


extern const BIG_256_56 CURVE_W_SECP256K1[2];	 /**< BN curve constant for GLV decomposition */
extern const BIG_256_56 CURVE_SB_SECP256K1[2][2]; /**< BN curve constant for GLV decomposition */
extern const BIG_256_56 CURVE_WB_SECP256K1[4];	 /**< BN curve constant for GS decomposition */
extern const BIG_256_56 CURVE_BB_SECP256K1[4][4]; /**< BN curve constant for GS decomposition */


/**
	@brief ECP structure - Elliptic Curve Point over base field
*/

typedef struct
{
//    int inf; /**< Infinity Flag - not needed for Edwards representation */

    FP_SECP256K1 x; /**< x-coordinate of point */
#if CURVETYPE_SECP256K1!=MONTGOMERY
    FP_SECP256K1 y; /**< y-coordinate of point. Not needed for Montgomery representation */
#endif
    FP_SECP256K1 z;/**< z-coordinate of point */
} ECP_SECP256K1;


/* ECP E(Fp) prototypes */
/**	@brief Tests for ECP point equal to infinity
 *
	@param P ECP point to be tested
	@return 1 if infinity, else returns 0
 */
extern int ECP_SECP256K1_isinf(ECP_SECP256K1 *P);
/**	@brief Tests for equality of two ECPs
 *
	@param P ECP instance to be compared
	@param Q ECP instance to be compared
	@return 1 if P=Q, else returns 0
 */
extern int ECP_SECP256K1_equals(ECP_SECP256K1 *P,ECP_SECP256K1 *Q);
/**	@brief Copy ECP point to another ECP point
 *
	@param P ECP instance, on exit = Q
	@param Q ECP instance to be copied
 */
extern void ECP_SECP256K1_copy(ECP_SECP256K1 *P,ECP_SECP256K1 *Q);
/**	@brief Negation of an ECP point
 *
	@param P ECP instance, on exit = -P
 */
extern void ECP_SECP256K1_neg(ECP_SECP256K1 *P);
/**	@brief Set ECP to point-at-infinity
 *
	@param P ECP instance to be set to infinity
 */
extern void ECP_SECP256K1_inf(ECP_SECP256K1 *P);
/**	@brief Calculate Right Hand Side of curve equation y^2=f(x)
 *
	Function f(x) depends on form of elliptic curve, Weierstrass, Edwards or Montgomery.
	Used internally.
	@param r BIG n-residue value of f(x)
	@param x BIG n-residue x
 */
extern void ECP_SECP256K1_rhs(FP_SECP256K1 *r,FP_SECP256K1 *x);

#if CURVETYPE_SECP256K1==MONTGOMERY
/**	@brief Set ECP to point(x,[y]) given x
 *
	Point P set to infinity if no such point on the curve. Note that y coordinate is not needed.
	@param P ECP instance to be set (x,[y])
	@param x BIG x coordinate of point
	@return 1 if point exists, else 0
 */
extern int ECP_SECP256K1_set(ECP_SECP256K1 *P,BIG_256_56 x);
/**	@brief Extract x coordinate of an ECP point P
 *
	@param x BIG on exit = x coordinate of point
	@param P ECP instance (x,[y])
	@return -1 if P is point-at-infinity, else 0
 */
extern int ECP_SECP256K1_get(BIG_256_56 x,ECP_SECP256K1 *P);
/**	@brief Adds ECP instance Q to ECP instance P, given difference D=P-Q
 *
	Differential addition of points on a Montgomery curve
	@param P ECP instance, on exit =P+Q
	@param Q ECP instance to be added to P
	@param D Difference between P and Q
 */
extern void ECP_SECP256K1_add(ECP_SECP256K1 *P,ECP_SECP256K1 *Q,ECP_SECP256K1 *D);
#else
/**	@brief Set ECP to point(x,y) given x and y
 *
	Point P set to infinity if no such point on the curve.
	@param P ECP instance to be set (x,y)
	@param x BIG x coordinate of point
	@param y BIG y coordinate of point
	@return 1 if point exists, else 0
 */
extern int ECP_SECP256K1_set(ECP_SECP256K1 *P,BIG_256_56 x,BIG_256_56 y);
/**	@brief Extract x and y coordinates of an ECP point P
 *
	If x=y, returns only x
	@param x BIG on exit = x coordinate of point
	@param y BIG on exit = y coordinate of point (unless x=y)
	@param P ECP instance (x,y)
	@return sign of y, or -1 if P is point-at-infinity
 */
extern int ECP_SECP256K1_get(BIG_256_56 x,BIG_256_56 y,ECP_SECP256K1 *P);
/**	@brief Adds ECP instance Q to ECP instance P
 *
	@param P ECP instance, on exit =P+Q
	@param Q ECP instance to be added to P
 */
extern void ECP_SECP256K1_add(ECP_SECP256K1 *P,ECP_SECP256K1 *Q);
/**	@brief Subtracts ECP instance Q from ECP instance P
 *
	@param P ECP instance, on exit =P-Q
	@param Q ECP instance to be subtracted from P
 */
extern void ECP_SECP256K1_sub(ECP_SECP256K1 *P,ECP_SECP256K1 *Q);
/**	@brief Set ECP to point(x,y) given just x and sign of y
 *
	Point P set to infinity if no such point on the curve. If x is on the curve then y is calculated from the curve equation.
	The correct y value (plus or minus) is selected given its sign s.
	@param P ECP instance to be set (x,[y])
	@param x BIG x coordinate of point
	@param s an integer representing the "sign" of y, in fact its least significant bit.
 */
extern int ECP_SECP256K1_setx(ECP_SECP256K1 *P,BIG_256_56 x,int s);

#endif

/**	@brief Multiplies Point by curve co-factor
 *
	@param Q ECP instance
 */
extern void ECP_SECP256K1_cfp(ECP_SECP256K1 *Q);

/**	@brief Maps random BIG to curve point of correct order
 *
	@param Q ECP instance of correct order
	@param w OCTET byte array to be mapped
 */
extern void ECP_SECP256K1_mapit(ECP_SECP256K1 *Q,octet *w);

/**	@brief Converts an ECP point from Projective (x,y,z) coordinates to affine (x,y) coordinates
 *
	@param P ECP instance to be converted to affine form
 */
extern void ECP_SECP256K1_affine(ECP_SECP256K1 *P);
/**	@brief Formats and outputs an ECP point to the console, in projective coordinates
 *
	@param P ECP instance to be printed
 */
extern void ECP_SECP256K1_outputxyz(ECP_SECP256K1 *P);
/**	@brief Formats and outputs an ECP point to the console, converted to affine coordinates
 *
	@param P ECP instance to be printed
 */
extern void ECP_SECP256K1_output(ECP_SECP256K1 * P);

/**	@brief Formats and outputs an ECP point to the console
 *
	@param P ECP instance to be printed
 */
extern void ECP_SECP256K1_rawoutput(ECP_SECP256K1 * P);

/**	@brief Formats and outputs an ECP point to an octet string
	The octet string is normally in the standard form 0x04|x|y
	Here x (and y) are the x and y coordinates in left justified big-endian base 256 form.
	For Montgomery curve it is 0x06|x
	If c is true, only the x coordinate is provided as in 0x2|x if y is even, or 0x3|x if y is odd
	@param c compression required, true or false
	@param S output octet string
	@param P ECP instance to be converted to an octet string
 */
extern void ECP_SECP256K1_toOctet(octet *S,ECP_SECP256K1 *P,bool c);
/**	@brief Creates an ECP point from an octet string
 *
	The octet string is normally in the standard form 0x04|x|y
	Here x (and y) are the x and y coordinates in left justified big-endian base 256 form.
	For Montgomery curve it is 0x06|x
	If in compressed form only the x coordinate is provided as in 0x2|x if y is even, or 0x3|x if y is odd
	@param P ECP instance to be created from the octet string
	@param S input octet string
	return 1 if octet string corresponds to a point on the curve, else 0
 */
extern int ECP_SECP256K1_fromOctet(ECP_SECP256K1 *P,octet *S);
/**	@brief Doubles an ECP instance P
 *
	@param P ECP instance, on exit =2*P
 */
extern void ECP_SECP256K1_dbl(ECP_SECP256K1 *P);
/**	@brief Multiplies an ECP instance P by a small integer, side-channel resistant
 *
	@param P ECP instance, on exit =i*P
	@param i small integer multiplier
	@param b maximum number of bits in multiplier
 */
extern void ECP_SECP256K1_pinmul(ECP_SECP256K1 *P,int i,int b);
/**	@brief Multiplies an ECP instance P by a BIG, side-channel resistant
 *
	Uses Montgomery ladder for Montgomery curves, otherwise fixed sized windows.
	@param P ECP instance, on exit =b*P
	@param b BIG number multiplier

 */
extern void ECP_SECP256K1_mul(ECP_SECP256K1 *P,BIG_256_56 b);
/**	@brief Calculates double multiplication P=e*P+f*Q, side-channel resistant
 *
	@param P ECP instance, on exit =e*P+f*Q
	@param Q ECP instance
	@param e BIG number multiplier
	@param f BIG number multiplier
 */
extern void ECP_SECP256K1_mul2(ECP_SECP256K1 *P,ECP_SECP256K1 *Q,BIG_256_56 e,BIG_256_56 f);
/**	@brief Get Group Generator from ROM
 *
	@param G ECP instance
 */
extern void ECP_SECP256K1_generator(ECP_SECP256K1 *G);


#endif
