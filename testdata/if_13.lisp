(defpurefun ((vanishes! :@loob) x) x)
(defpurefun ((force-bin :binary) x) x)

(defcolumns (A :@loob) B C)
(defconstraint c1 ()
  (if (vanishes! (force-bin A))
      (vanishes! B)
      (vanishes! C)))
