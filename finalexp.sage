#!/usr/bin/env sage

import sys
from sage.all import *

# example terminal comment:
# ./finalexp.sage 9586122913090633729 6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299 258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177 -1 1 1 1 0 1 0 1 0 1 0 1 0 1 0

if len(sys.argv) < 7:
    print >> sys.stderr, "Usage: %s <t> <p> <r> <a> <b> <points>..." % sys.argv[0]
    print >> sys.stderr, "Outputs the sum, product, etc of <points> in the 3-over-2 tower"
    sys.exit(1)

# BLS12 parameter t
t=sage_eval(sys.argv[1])

 # fp field
p=sage_eval(sys.argv[2]) # prime characteristic p
fp=GF(p)
print "p:", p

# subgroup order r
r=sage_eval(sys.argv[3])
print "r:", r

# final exponent: (p^6-1)/r
easy_part = (p**3 - 1)*(p+1)
hard_part = (p**2 - p + 1)/r
hard_part_multiple = 3*(t**3 - t**2 + 1)
exponent = easy_part * hard_part * hard_part_multiple
# print "exponent:", exponent

# fp2 field
a=fp(sage_eval(sys.argv[4])) # a must be a quadratic nonresidue modulo p
print "a:", a

# fp6 field
b0=fp(sage_eval(sys.argv[5]))
b1=fp(sage_eval(sys.argv[6]))
print "b:", b0, " + ", b1, "u"
if b1.is_zero():
    print >> sys.stderr, "coefficient of u in b cannot be 0"
    sys.exit(1)

# fp6 field irrep
P.<x>=PolynomialRing(fp)
fp6_modulus = P(x^6 - (b0^2 + a*b1^2))
if not b0.is_zero():
    fp6_modulus -= P(2*b0*x^3 - 2*b0^2)
# print "fp6 modulus:", fp6_modulus

fp6.<v>=GF(p^6, modulus=fp6_modulus)

# matrix embedding fp2 into fp6
M = Matrix(fp, [[1, -b0*b1^(-1)], [0, b1]])
# print "M:\n", M
M_inverse = M.inverse()
# print "M_inverse:\n", M_inverse

def print_result(out):

    # python truncates leading 0s
    outlist = out.polynomial().list()
    while len(outlist) < 6:
        outlist.append(0)

    # unpack out0, out1, out2 and map back into fp2
    out0 = M_inverse * vector(fp, [outlist[0], outlist[3]])
    out1 = M_inverse * vector(fp, [outlist[1], outlist[4]])
    out2 = M_inverse * vector(fp, [outlist[2], outlist[5]])

    print out0[0]
    print out0[1]
    print out1[0]
    print out1[1]
    print out2[0]
    print out2[1]

def FinalExp_easy_BW6_761(m):
    mp3 = m.frobenius(3) # assuming the extension is in one layer
    im = 1/m
    f = mp3*im; # m^(q^3-1)
    f = f * f.frobenius() # m^((q^3-1)*(q+1))
    return f

def FinalExp_hard_BW6_761_multi_2NAF(m,u0):
    """ returns m^exponent where exponent is a multiple of (q^2-q+1)/r"""
    f0 = m
    f1 = m**u0
    f2 = f1**u0
    f3 = f2**u0
    f4 = f3**u0
    f5 = f4**u0
    f6 = f5**u0
    f7 = f6**u0
    f0p = m.frobenius()
    f1p = f1.frobenius()
    f2p = f2.frobenius()
    f3p = f3.frobenius()
    f4p = f4.frobenius()
    f5p = f5.frobenius()
    f6p = f6.frobenius()
    f7p = f7.frobenius()
    f8p = f7p**u0
    f9p = f8p**u0

    f = f3p*f6p*(f5p).frobenius(3)                         # 2M
    f = f**2
    f4f2p = f4*f2p                                         # 1M
    f *= f5*f0p*(f0*f1*f3*f4f2p*f8p).frobenius(3)          # 7M
    f = f**2
    f *= f9p*(f7).frobenius(3)                             # 2M
    f = f**2
    f2f4p = f2*f4p                                         # 1M
    f4f2pf5p = f4f2p*f5p                                   # 1M
    f *= f4f2pf5p*f6*f7p*(f2f4p*f3*f3p).frobenius(3)       # 6M
    f = f**2
    f *= f0*f7*f1p*(f0p*f9p).frobenius(3)                  # 5M
    f = f**2
    f6pf8p = f6p*f8p                                       # 1M
    f5f7p = f5*f7p                                         # 1M
    f *= f5f7p*f2p*(f6pf8p).frobenius(3)                   # 3M
    f = f**2
    f3f6 = f3*f6                                           # 1M
    f1f7 = f1*f7                                           # 1M
    f *= f3f6*f9p*(f1f7*f2).frobenius(3)                   # 4M
    f = f**2
    f *= f0*f0p*f3p*f5p*(f4f2p*f5f7p*f6pf8p).frobenius(3)  # 7M
    f = f**2
    f *= f1p*(f3f6).frobenius(3)                           # 2M
    f = f**2
    f *= f1f7*f5f7p*f0p*(f2f4p*f4f2pf5p*f9p).frobenius(3)  # 6M
    # 51 M
    #  9 S
    return f

def FinalExp_BW6_761(m,u0):
    f = FinalExp_easy_BW6_761(m)
    #f = FinalExp_hard_BW6_761(f,u0)
    f = FinalExp_hard_BW6_761_multi_2NAF(f,u0)
    return f

for i in range(7, len(sys.argv), 12):

    # parse in1, in2
    # in00, in01, in02 are in fp2, need to be embedded into fp6
    in00 = M * vector(fp, [sage_eval(sys.argv[i]), sage_eval(sys.argv[i+1])])
    in01 = M * vector(fp, [sage_eval(sys.argv[i+2]), sage_eval(sys.argv[i+3])])
    in02 = M * vector(fp, [sage_eval(sys.argv[i+4]), sage_eval(sys.argv[i+5])])
    in0 = fp6([in00[0], in01[0], in02[0], in00[1], in01[1], in02[1]])

    in10 = M * vector(fp, [sage_eval(sys.argv[i+6]), sage_eval(sys.argv[i+7])])
    in11 = M * vector(fp, [sage_eval(sys.argv[i+8]), sage_eval(sys.argv[i+9])])
    in12 = M * vector(fp, [sage_eval(sys.argv[i+10]), sage_eval(sys.argv[i+11])])
    in1 = fp6([in10[0], in11[0], in12[0], in10[1], in11[1], in12[1]])

    # binary ops
    # print_result(in0+in1) # add
    # print_result(in0-in1) # sub
    # print_result(in0*in1) # mul
    # # print_result(in0*fp6([in10[0], 0, 0, in10[1], 0, 0])) # mul by fp2 element

    # # unary ops ignore in1
    # # print_result(in0*fp6(v)) # mul by gen (ie. mul by v=(0,1,0) in fp6)
    # print_result(in0*in0) # square

    # # inv
    # if in0==fp6.zero():
    #     print_result(fp6.zero()) # can't invert 0; just output 0
    # else:
    #     print_result(in0^(-1))

    # # print_result(in0.conjugate()) # conjugate
    # print_result(fp6.frobenius_endomorphism()(in0)) # frobenius
    # print_result(fp6.frobenius_endomorphism(2)(in0)) # frobenius squared
    # print_result(fp6.frobenius_endomorphism(3)(in0)) # frobenius cubed
    # print_result(in0^t) # expt
    # print_result(in0^exponent) # final exponent
    u0=0x8508C00000000001
    print "t right?", u0 == t
    result1 = in0^easy_part
    result2 = FinalExp_easy_BW6_761(in0)
    print "easy part?", result1 == result2
    result1 = in0^(hard_part * hard_part_multiple)
    result2 = FinalExp_hard_BW6_761_multi_2NAF(in0, t)
    print "hard part?", result1 == result2
    # print "me:"
    # print_result(result1)
    # print "them:"
    # print_result(result2)
    result1 = in0^exponent
    result2 = FinalExp_BW6_761(in0, t)
    print "final?", result1 == result2