(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (A :i16@loob) B C)
(defconstraint c1 ()
  (if A
      (vanishes! B)
      (vanishes! C)))
