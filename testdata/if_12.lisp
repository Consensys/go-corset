(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (A :binary@loob) (B :i16) (C :i16))
(defconstraint c1 ()
  (if A
      (vanishes! B)
      (vanishes! C)))
