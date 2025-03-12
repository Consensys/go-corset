(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (A :i16@loob) (B :i16))
(defconstraint c1 ()
  (if A
      (vanishes! 0)
      (vanishes! B)))
