(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (ST :i5) (A :i16@loob) (B :i16) (C :i16))
(defconstraint c1 (:guard ST)
  (if A
      (vanishes! B)
      (vanishes! C)))
