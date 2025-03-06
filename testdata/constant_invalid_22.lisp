;;error:9:23-26:not permitted in const context
;;error:9:53-56:not permitted in const context
(defpurefun ((vanishes! :@loob) x) x)

(defconst (ONE :extern) 1)
(defcolumns CT)
(defconstraint c1 ()
  (vanishes!
   (* (- CT (shift CT ONE)) (- (+ CT ONE) (shift CT ONE)))))
