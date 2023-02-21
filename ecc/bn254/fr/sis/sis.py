r = 21888242871839275222246405745257275088548364400416034343698204186575808495617
Fr = GF(r)
Fr.<x> = Fr[]

# Montgomery constant
rr = Fr(2**256)

# utils
def buildPoly(a):
    """ Builds a poly from the array a

    Args:
        a an array

    Returns:
        a[0]+a[1]*X + .. + a[n]*X**n
    """

    res = 0
    for i, v in enumerate(a):
        res += v*x**i
    return res

def bitAt(i, b):
    """
    Args:
        i: index of the bit to retrieve
        b: array of bytes

    Returns:
        the i-th bit of b, when it is written b[0] || b[1] || ...
    """
    k = i//8
    j = i%8
    return (b[k] >> (7-j)) & 1

def toBytes(m, s):
    """

    Args:
        m: a bit int
        s: the expected number of bytes of the result. If s is bigger than the
        number of bytes in m, the remaining bytes are set to zero.
    
    Returns:
        the byte representation of m as a byte array, as
        in gnark-crypto.
    """
    res = s*[0]
    mask = 255
    for i in range(s):
        res[s-1-i] = m & mask
        m = m>>8
    return res

def splitCoeffs(b, logTwoBound):
    """
    Args:
        b: an array of bytes
        logTwoBound: number of bits of the bound

    Returns:
        an array of coeffs, each coeff being the i-th chunk of logTwoBounds bits of b.
    """
    nbBits = len(b)*8
    nbCoeffs = nbBits // logTwoBound # remainder is supposed to be zero
    res = nbCoeffs * [0]
    p = 0
    i = 0
    while i<nbBits:
        for j in range(logTwoBound):
            res[p] += bitAt(i, b)<<j
            i+=1
        p+=1
    return [Fr(res[i])*rr**-1 for i in range(nbCoeffs)] # careful Montgomery constant...


# pseudo random generators
def pRand(seed):
    return seed**2

def polyRand(seed, n):
    """ Generates a pseudo random polynomial of size n from seed.

    Args:
        seed: seed for the pseudo random gen
        n: degree of the polynomial
    """
    
    a = n*[0]
    for i in range(n):
        a[i] = pRand(seed)
        seed = a[i]
    return buildPoly(a)


# SIS
class Sis:
    def __init__(self, seed, size, degree, logTwoBound):
        """
            Args:
                size: size of the key
                degree: degree of the polynomials
                logTwoBound: bound of SIS
        """
        self.logTwoBound = logTwoBound
        self.degree = degree
        self.size = size
        self.key = size * [0]
        for i in range(size):
            self.key[i] = polyRand(seed, self.degree)
            seed+=1


    def hash(self, b):
        """ 
        Args:
            b is a list of bytes to hash

        Returns:
            the sis hash of m.
        """
        # step 1: build the polynomials from m
        c = splitCoeffs(b, self.logTwoBound) 
        mp = [buildPoly(c[self.degree*i:self.degree*(i+1)]) for i in range(self.size)]

        # step 2: compute sum_i mp[i]*key[i] mod X^n+1
        modulo = x**self.degree+1
        res = 0
        for i in range(self.size):
            res += self.key[i]*mp[i]
        res = res % modulo
        return res

