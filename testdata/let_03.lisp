(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (X :@loob) (Y :@loob) Z)
(defconstraint test ()
  (let ((THREE 3))
    (if X
        (vanishes! 0)
        (vanishes! (- Z (if Y THREE 16))))))
