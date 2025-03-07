(defpurefun ((vanishes! :𝔽@loob) x) x)
(defpurefun ((force-bin :binary) x) x)

(defcolumns (A :i16@loob) B C)
(defconstraint c1 ()
  (if (vanishes! (force-bin A))
      (vanishes! B)
      (vanishes! C)))
