# https://eprint.iacr.org/2023/323.pdf, page 8

#https://github.com/argumentcomputer/neptune/blob/main/spec/poseidon_spec.pdf

r_F = 8
r_f = r_F/2

# security
# M ∈ {80,128,256}, security level in bits
# cf https://github.com/argumentcomputer/neptune/blob/main/spec/poseidon_spec.pdf

# 2^M < p^t (M < t log_2 p)

# Rf > 6
# eq 2 section 5.5.1, log_2(p)-2=252
# RF >  6 if M < (log_2(p)-C)(t+1) where C = log_2(alpha-1), alpha=sbox's exponent
#       10 otherwise

# R = RF + RP (number of total full rounds and partial rounds)
# R > log_alpha(2)*min{M, log_2(p)} + log_alpha(t) (condition 1)
# R > (M log_alpha(2)) / 3 (condition 2)
# R > t-1  + (M log_alpha(2))/(t+1) (condition 3)

def find_t(M, p):
    """
    2^M < p^t so t must be at least M / log_2 p

    Args:
        M: number of bits of security
        p: field of definition
    """
    return float(M)/(log(float(p))/log(2.))

def find_r_cond_1(alpha, M, p, t):
    """
    log_alpha(2)*min{M, log_2(p)} + log_alpha(t)

    Args:
        alpha: degree of the sbox
        M: number of bits of security
        p: field of definition
        t: size of the block
    """
    a = (log(2.)/log(float(alpha)))*float(M)
    b = (log(2.)/log(float(alpha))*log(float(p))/log(2.)
    c = log(float(t))/log(float(alpha))
    d = min(a,b)+c
    return d

def find_r_cond_2(alpha, M):
    """
    M*log_alpha(2)/3

    Args:
        alpha: degree of the sbox
        M: number of bits of security

    """
    return M * log(2.)/(3.*log(float(alpha)))

def find_r_cond_3(alpha, M, t):
    """
    t-1+ (M log_alpha(2)/(t+1))
    
    Args:
        alpha: degree of the sbox
        M: number of bits of security
        t: size of the block
    """
    a = float(t-1)
    b = M * log(2.)/(log(float(alpha))*float(t+1))
    return a + b

def find_r(alpha, M, p, t):
    """
    finds the number of rounds R, satisying the following inequalities:
    R > log_alpha(2)*min{M, log_2(p)} + log_alpha(t) (condition 1)
    R > (M log_alpha(2)) / 3 (condition 2)
    R > t-1  + (M log_alpha(2))/(t+1) (condition 3)

    Args:
        alpha: degree of the sbox
        M: number of bits of security
        p: field of definition
        t: size of the block
    """
    a = find_r_cond_1(alpha, M, p, 1)
    b = find_r_cond_2(alpha, M)
    c = find_r_cond_3(alpha, M, t)
    a = max(a, b)
    a = max(a, c)
    return a


def find_degree_sbox(p):
    """
    find the degree of the sbox

    Args:
        p: prime field on which variables live
    """
    d = 2
    while (p-1)%d==0:
        d = d + 1
    return d

def min(a, b):
    if a<b:
        return a
    else:
        return b

def max(a, b):
    if a<b:
        return b
    else:
        return a

def generate_image(V, m, sp):
    """
    returns the subspace m*sp
    
    Args:
        V: ambiant vector space
        m: matrix
        sp: the sub vector space on which m acts
    """
    new_basis = []
    for vec in sp.basis():
        v = m*vec
        new_basis = new_basis + [v]
    return V.subspace(new_basis)

def generate_matrix_internal_round(F, t):
    """
    Returns a matrix full of ones, except on the diagonal,
    whose minimal polynomial is irreducible, so that the Frobenius
    normal form

    Args:
        F: finite field
        t: size of matrix (t x t)
    """
    res = []
    for i in range(t):
        tmp = t*[Fp(1)]
        res = res + [tmp]
        res[i][i] = F.random_element()
    m = matrix(res)
    mp = m.minimal_polynomial()
    while mp.is_irreducible==True:
        for i in range(t):
            res[i][i] = F.random_element()
        m = matrix(res)
    return m

def generate_subspace_wo_active_sbox(V, s):
    """
    From the vector space V of dimension n, generates the subspace spanned
    by e_s+1, .., e_n.

    Args:
        V: ambient vector space
        s: number of inactive sboxes
    """
    res = []
    basis = V.basis()
    return V.subspace(basis[s:])

def find_stable_image(V, sp, m):
    """
    Returns the subspace which is the intersection of (sp,m.sp,m^2.sp, ...)

    Args:
        V: ambient vector space
        sp: subspace
        m: matrix
    """
    sp_next = generate_image(V, m, sp)
    while sp_next!=sp and sp.dimension()!=0:
        sp = sp_next.intersection(sp)
        sp_next = generate_image(V, m, sp)
    return sp


def algo_1(Fp, V, s, m):
    """
    Applies algorithm 1 of https://tosc.iacr.org/index.php/ToSC/article/view/8913/8489.
    The matrix found in this function has also the property that its minimal polynomial
    is irreducible, so there is only one invariant subspace, equivalently the Frobenius
    normal form of the matrix contains one block.

    Args:
        Fp: field of definition
        V: ambient vector space
        s: number of inactive sboxes
        m: matrix to test
    """
    t = V.dimension()
    sp = generate_subspace_wo_active_sbox(V, s)

    m = generate_matrix_internal_round(Fp, t)
    stable_space = find_stable_image(V, sp, m)
    
    return stable_space.dimension()==0 

def algo_2(Fp, V, s, m):
    """
    Returns True if an infinitely long subspace trail has been found.

    Args:
        Fp: base field
        V: ambient vector space
        s: number of active sboxes
        m: matrix to test
    """
    basis = V.basis()
    subsets = powerset(range(s))

    complement = []
    for i in range(s, V.dimension()):
        complement = complement + [basis[i]]

    res = False
    
    for I_s in subsets:

        if res == True:
            break

        I = []
 
        for i in I_s:
            I = I + [basis[i]]
        
        complement_and_I = complement + I
        complement_and_I = V.subspace(complement_and_I)
        
        I = V.subspace(I)
        
        next_I_s = False

        for i in I_s:
           
            v = basis[i]

            delta = I.dimension()
            v = m*v
            I_basis = I.basis()
            I_basis = I_basis + [v]
            I = V.subspace(I_basis)

            
            while I.dimension() > delta:
                
                if I.dimension()==t or I.intersection(complement_and_I)!=I:
                    next_I_s = True
                    break

                delta = I.dimension()
                v = m*v
                I_basis = I.basis()
                I_basis = I_basis + [v]
                I = V.subspace(I_basis)

            if next_I_s == True:
                break

            if next_I_s==False:
                res = True

    return res          

# M4 = 
#(5 7 1 3)
#(4 6 1 1)
#(1 3 5 7)
#(1 1 4 6)
