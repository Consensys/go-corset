(defpurefun ((vanishes! :𝔽@loob :force) e0) e0)
(defpurefun ((eq! :𝔽@loob) x y) (- x y))

(module m1)

(defcolumns
    (ACC_1 :i128)
    (BYTE :byte :array [0:2])
)

(defconstraint c1 () (vanishes! (if (eq! ACC_1 1) [BYTE 0])))
