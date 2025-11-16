(module sq)
(defcolumns (X_lo :i128) (X_hi :i128) (Y_lo :i128) (Y_hi :i128))
(defun (X) (:: X_hi X_lo))
(defun (Y) (:: Y_hi Y_lo))
;; Y = (X + X) % 2^256
(defcall ((Y)) add ((X) (X)))
