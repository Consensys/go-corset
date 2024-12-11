(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (X :@loob) (Y :@loob) Z)
(defconstraint test ()
  (if X
      (vanishes! 0)
      (vanishes! (- Z (if Y 3 16)))))
