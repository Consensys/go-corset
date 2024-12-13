(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (A :@loob) B C)
(defconstraint c1 ()
  (if A
      (vanishes! B)
      (vanishes! C)))
