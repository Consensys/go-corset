(defcolumns (BIT_1 :binary@loob@prove) (ARG :@loob))

(defconstraint pivot ()
        ;; If BIT_1[k-1]=0 and BIT_1[k]=1
        (if (+ (shift BIT_1 -1) (- 1 BIT_1))
            ;; Then ARG = 0
            ARG))
