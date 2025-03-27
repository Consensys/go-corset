(defpurefun ((vanishes! :𝔽@loob :force) e0) e0)
(defpurefun ((eq! :𝔽@loob) x y) (- x y))

(module m1)
(defcolumns
    (ACC_1 :i128)
    (BYTE :byte :array [0:1])
)
(defconstraint test () (vanishes! (if (eq! ACC_1 1) [BYTE 0])))

(module m2)
(defcolumns (A :i128) (B :byte))
(deflookup
  l1
  ;; target columns
  (m1.ACC_1 [m1.BYTE 1])
  ;; source columns
  (A B))
