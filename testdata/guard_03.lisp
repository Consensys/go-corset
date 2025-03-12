(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (ST :i3) (A :i16@loob) (B :i16))
(defconstraint c1 (:guard ST)
  (if A
      (vanishes! 0)
      (vanishes! B)))
